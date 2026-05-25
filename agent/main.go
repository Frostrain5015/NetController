package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"agent/collector"
	"agent/config"
	"agent/ws"
)

type Snapshot struct {
	Timestamp     int64                   `json:"timestamp"`
	ServerMetrics collector.SystemMetrics `json:"serverMetrics"`
	Projects      []ProjectStatus         `json:"projects"`
	Proxy         ProxyStatus             `json:"proxy"`
	ProxyNodes    []ProxyNodeOut          `json:"proxyNodes"`
}

type ProjectStatus struct {
	Name     string  `json:"name"`
	Alive    bool    `json:"alive"`
	PID      int32   `json:"pid"`
	Port     int     `json:"port"`
	PortOpen bool    `json:"portOpen"`
	CPU      float64 `json:"cpuPercent"`
	MemMB    uint64  `json:"memMB"`
}

type ProxyStatus struct {
	Alive              bool    `json:"alive"`
	Port               int     `json:"port"`
	PortOpen           bool    `json:"portOpen"`
	ActiveConnections  int     `json:"activeConnections"`
	ApiAccessible      bool    `json:"apiAccessible"`
	TrafficRemainingGB float64 `json:"trafficRemainingGB"`
	PlanExpiry         string  `json:"planExpiry"`
}

type ProxyNodeOut struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Group       string    `json:"group"`
	Type        string    `json:"type"`
	LatencyMs   int       `json:"latencyMs"`
	Reachable   bool      `json:"reachable"`
	Location    []float64 `json:"location"`
	Country     string    `json:"country"`
}

func main() {
	cfgPath := flag.String("c", "agent.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	hub := ws.NewHub()
	http.HandleFunc("/ws", hub.HandleWS)

	go func() {
		log.Printf("agent listening on %s", cfg.Listen)
		if err := http.ListenAndServe(cfg.Listen, nil); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	// 首次采集
	collectAndBroadcast(hub, cfg)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ticker.C:
			collectAndBroadcastSys(hub, cfg)
		case <-pingTicker.C:
			pingAllNodes(hub, cfg)
		}
	}
}

var cachedNodes []proxyNodeInfo
var prevReachable = make(map[string]bool)
var prevLatencies = make(map[string]int)
var apiPort int
var testURL string
var cachedTrafficGB float64
var cachedExpiry string

func collectProjectStatus(cfg *config.Config) []ProjectStatus {
	var projects []ProjectStatus
	for _, p := range cfg.Projects {
		pid, alive := collector.FindProcessByName(p.ProcessName)
		cpu, memMB := collector.GetProcessCPUAndMem(p.ProcessName)
		projects = append(projects, ProjectStatus{
			Name: p.Name, Alive: alive, PID: pid,
			Port: p.Port, PortOpen: collector.CheckPort(p.Port),
			CPU: cpu, MemMB: memMB,
		})
	}
	return projects
}

func collectAndBroadcast(hub *ws.Hub, cfg *config.Config) {
	snap := Snapshot{
		Timestamp:     time.Now().Unix(),
		ServerMetrics: collector.CollectSystem(),
		Projects:      collectProjectStatus(cfg),
	}

	_, pxAlive := collector.FindProcessByName(cfg.Proxy.ProcessName)
	apiPort = cfg.Proxy.ClashApiPort
	if apiPort == 0 {
		apiPort = 9090
	}
	testURL = cfg.Proxy.TestURL
	if testURL == "" {
		testURL = "https://www.gstatic.com/generate_204"
	}
	snap.Proxy = ProxyStatus{
		Alive:   pxAlive,
		Port:    cfg.Proxy.Port,
		PortOpen: collector.CheckPort(cfg.Proxy.Port),
	}

	if pxAlive {
		nodes, trafficGB, expiry := fetchClashProxies(apiPort)
		cachedTrafficGB = trafficGB
		cachedExpiry = expiry
		snap.Proxy.TrafficRemainingGB = trafficGB
		snap.Proxy.PlanExpiry = expiry
		if len(nodes) > 0 {
			snap.Proxy.ApiAccessible = true
			cachedNodes = nodes
			snap.ProxyNodes = make([]ProxyNodeOut, 0, len(nodes))
			for _, n := range nodes {
				snap.ProxyNodes = append(snap.ProxyNodes, ProxyNodeOut{
					Name: n.Name, DisplayName: n.DisplayName,
					Group: n.Group, Type: n.Type,
					Location: n.Location, Country: n.Country,
				})
			}
		}
	}

	data, _ := json.Marshal(snap)
	hub.SetSnapshot(data)
	hub.Broadcast(data)

	if len(cachedNodes) > 0 {
		go pingAllNodes(hub, cfg)
	}
}

func collectAndBroadcastSys(hub *ws.Hub, cfg *config.Config) {
	snap := Snapshot{
		Timestamp:     time.Now().Unix(),
		ServerMetrics: collector.CollectSystem(),
		Projects:      collectProjectStatus(cfg),
	}
	_, pxAlive := collector.FindProcessByName(cfg.Proxy.ProcessName)
	snap.Proxy = ProxyStatus{
		Alive: pxAlive, Port: cfg.Proxy.Port,
		PortOpen: collector.CheckPort(cfg.Proxy.Port),
		ApiAccessible:      len(cachedNodes) > 0,
		TrafficRemainingGB: cachedTrafficGB,
		PlanExpiry:         cachedExpiry,
	}
	data, _ := json.Marshal(snap)
	hub.Broadcast(data)
}

func pingAllNodes(hub *ws.Hub, cfg *config.Config) {
	log.Printf("pinging %d proxy nodes via :%d (testURL=%s)...", len(cachedNodes), apiPort, testURL)
	var wg sync.WaitGroup
	next := make(map[string]bool)
	var mu sync.Mutex

	for _, n := range cachedNodes {
		wg.Add(1)
		go func(node proxyNodeInfo) {
			defer wg.Done()
			ms := probeSingleDelay(apiPort, node.Name, testURL, 1)
			reachable := ms > 0 && ms <= 1000
			was := prevReachable[node.Name]
			mu.Lock()
			next[node.Name] = reachable
				prevLatencies[node.Name] = ms
			mu.Unlock()
			msg, _ := json.Marshal(map[string]interface{}{
				"type":         "proxy-ping-result",
				"name":         node.Name,
				"latencyMs":    ms,
				"reachable":    reachable,
				"wasReachable": was,
			})
			hub.Broadcast(msg)
		}(n)
	}
	wg.Wait()
	prevReachable = next
	log.Printf("ping cycle complete: %d reachable", reachableCount(next))

	// 更新缓存快照，使重连客户端立即获取当前状态
	if len(cachedNodes) > 0 {
		snap := Snapshot{
			Timestamp:     time.Now().Unix(),
			ServerMetrics: collector.CollectSystem(),
			Projects:      collectProjectStatus(cfg),
		}
		snap.Proxy = ProxyStatus{
			Alive: true, Port: cfg.Proxy.Port,
			PortOpen: collector.CheckPort(cfg.Proxy.Port),
			ApiAccessible:      true,
			TrafficRemainingGB: cachedTrafficGB,
			PlanExpiry:         cachedExpiry,
		}
		for _, n := range cachedNodes {
			reachable := prevReachable[n.Name]
			latency := prevLatencies[n.Name]
			snap.ProxyNodes = append(snap.ProxyNodes, ProxyNodeOut{
				Name: n.Name, DisplayName: n.DisplayName,
				Group: n.Group, Type: n.Type,
				Location: n.Location, Country: n.Country,
				Reachable: reachable, LatencyMs: latency,
			})
		}
		data, _ := json.Marshal(snap)
		hub.SetSnapshot(data)
		hub.Broadcast(data) // 推送更新到所有已连接客户端
	}
}

func reachableCount(m map[string]bool) int {
	c := 0
	for _, v := range m {
		if v {
			c++
		}
	}
	return c
}
