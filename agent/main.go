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
	Alive              bool     `json:"alive"`
	Port               int      `json:"port"`
	PortOpen           bool     `json:"portOpen"`
	ActiveConnections  int      `json:"activeConnections"`
	ApiAccessible      bool     `json:"apiAccessible"`
	TrafficRemainingGB *float64 `json:"trafficRemainingGB"`
	TrafficUsedGB      *float64 `json:"trafficUsedGB"`
	TrafficTotalGB     *float64 `json:"trafficTotalGB"`
	PlanExpiry         string   `json:"planExpiry"`
}

type ProxyNodeOut struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Group       string    `json:"group"`
	Type        string    `json:"type"`
	GroupType   string    `json:"groupType"`
	LatencyMs   int       `json:"latencyMs"`
	Reachable   bool      `json:"reachable"`
	Selected    bool      `json:"selected"`
	Location    []float64 `json:"location"`
	Country     string    `json:"country"`
}

type ClientMessage struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Group string `json:"group"`
}

func main() {
	cfgPath := flag.String("c", "agent.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	hub := ws.NewHub()
	hub.SetMessageHandler(func(msg []byte) {
		handleClientMessage(hub, cfg, msg)
	})
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
var cachedTrafficRemainingGB *float64
var cachedTrafficUsedGB *float64
var cachedTrafficTotalGB *float64
var cachedExpiry string

var deprecatedProjects = map[string]bool{
	"Webhook":        true,
	"deploy-webhook": true,
}

func collectProjectStatus(cfg *config.Config) []ProjectStatus {
	var projects []ProjectStatus
	for _, p := range cfg.Projects {
		if deprecatedProjects[p.Name] || deprecatedProjects[p.ProcessName] {
			continue
		}
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
		Alive:    pxAlive,
		Port:     cfg.Proxy.Port,
		PortOpen: collector.CheckPort(cfg.Proxy.Port),
	}

	if pxAlive {
		nodes, usage := fetchClashProxies(apiPort, cfg.Proxy.SubscriptionURL)
		updateProxyCache(nodes, usage)
		snap.Proxy.ActiveConnections = fetchClashConnectionCount(apiPort)
		snap.Proxy.TrafficRemainingGB = cachedTrafficRemainingGB
		snap.Proxy.TrafficUsedGB = cachedTrafficUsedGB
		snap.Proxy.TrafficTotalGB = cachedTrafficTotalGB
		snap.Proxy.PlanExpiry = cachedExpiry
		if len(nodes) > 0 {
			snap.Proxy.ApiAccessible = true
			snap.ProxyNodes = make([]ProxyNodeOut, 0, len(nodes))
			for _, n := range nodes {
				snap.ProxyNodes = append(snap.ProxyNodes, proxyNodeOutWithStatus(n))
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
	}
	if pxAlive {
		nodes, usage := fetchClashProxies(apiPort, cfg.Proxy.SubscriptionURL)
		updateProxyCache(nodes, usage)
		snap.Proxy.ActiveConnections = fetchClashConnectionCount(apiPort)
		snap.Proxy.ApiAccessible = len(cachedNodes) > 0
		snap.Proxy.TrafficRemainingGB = cachedTrafficRemainingGB
		snap.Proxy.TrafficUsedGB = cachedTrafficUsedGB
		snap.Proxy.TrafficTotalGB = cachedTrafficTotalGB
		snap.Proxy.PlanExpiry = cachedExpiry
		for _, n := range cachedNodes {
			snap.ProxyNodes = append(snap.ProxyNodes, proxyNodeOutWithStatus(n))
		}
	} else {
		updateProxyCache(nil, subscriptionUsage{})
	}
	data, _ := json.Marshal(snap)
	hub.SetSnapshot(data)
	hub.Broadcast(data)
}

func updateProxyCache(nodes []proxyNodeInfo, usage subscriptionUsage) {
	cachedNodes = nodes
	cachedTrafficRemainingGB = usage.RemainingGB
	cachedTrafficUsedGB = usage.UsedGB
	cachedTrafficTotalGB = usage.TotalGB
	cachedExpiry = usage.Expiry
}

func proxyNodeOut(n proxyNodeInfo) ProxyNodeOut {
	return ProxyNodeOut{
		Name: n.Name, DisplayName: n.DisplayName,
		Group: n.Group, GroupType: n.GroupType, Type: n.Type,
		Location: n.Location, Country: n.Country, Selected: n.Selected,
	}
}

func proxyNodeOutWithStatus(n proxyNodeInfo) ProxyNodeOut {
	out := proxyNodeOut(n)
	out.Reachable = prevReachable[n.Name]
	out.LatencyMs = prevLatencies[n.Name]
	return out
}

func handleClientMessage(hub *ws.Hub, cfg *config.Config, data []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}
	if msg.Type != "proxy-select" || msg.Name == "" {
		return
	}
	port := apiPort
	if port == 0 {
		port = cfg.Proxy.ClashApiPort
	}
	if port == 0 {
		port = 9090
	}
	err := selectProxyNode(port, msg.Group, msg.Name)
	result := map[string]interface{}{
		"type":  "proxy-select-result",
		"name":  msg.Name,
		"group": msg.Group,
		"ok":    err == nil,
	}
	if err != nil {
		result["message"] = err.Error()
	}
	payload, _ := json.Marshal(result)
	hub.Broadcast(payload)
	if err == nil {
		collectAndBroadcast(hub, cfg)
	}
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
			PortOpen:           collector.CheckPort(cfg.Proxy.Port),
			ApiAccessible:      true,
			ActiveConnections:  fetchClashConnectionCount(apiPort),
			TrafficRemainingGB: cachedTrafficRemainingGB,
			TrafficUsedGB:      cachedTrafficUsedGB,
			TrafficTotalGB:     cachedTrafficTotalGB,
			PlanExpiry:         cachedExpiry,
		}
		for _, n := range cachedNodes {
			snap.ProxyNodes = append(snap.ProxyNodes, proxyNodeOutWithStatus(n))
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
