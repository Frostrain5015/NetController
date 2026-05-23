package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"agent/collector"
	"agent/config"
	"agent/ws"
)

type Snapshot struct {
	Timestamp     int64                    `json:"timestamp"`
	ServerMetrics collector.SystemMetrics  `json:"serverMetrics"`
	Projects      []ProjectStatus          `json:"projects"`
	Proxy         ProxyStatus              `json:"proxy"`
	OverseasNodes []collector.ProbeResult  `json:"overseasNodes"`
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
	Alive             bool   `json:"alive"`
	Port              int    `json:"port"`
	PortOpen          bool   `json:"portOpen"`
	ActiveConnections int    `json:"activeConnections"`
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

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		snap := collect(cfg)
		data, err := json.Marshal(snap)
		if err != nil {
			log.Printf("marshal: %v", err)
			continue
		}
		hub.Broadcast(data)
	}
}

func collect(cfg *config.Config) Snapshot {
	snap := Snapshot{
		Timestamp:     time.Now().Unix(),
		ServerMetrics: collector.CollectSystem(),
	}

	for _, p := range cfg.Projects {
		pid, alive := collector.FindProcessByName(p.ProcessName)
		ps := ProjectStatus{
			Name:     p.Name,
			Alive:    alive,
			PID:      pid,
			Port:     p.Port,
			PortOpen: collector.CheckPort(p.Port),
		}
		snap.Projects = append(snap.Projects, ps)
	}

	_, pxAlive := collector.FindProcessByName(cfg.Proxy.ProcessName)
	snap.Proxy = ProxyStatus{
		Alive:    pxAlive,
		Port:     cfg.Proxy.Port,
		PortOpen: collector.CheckPort(cfg.Proxy.Port),
	}

	for _, n := range cfg.OverseasNodes {
		reachable, latency := collector.ProbeTCP(n.Host, 5*time.Second)
		snap.OverseasNodes = append(snap.OverseasNodes, collector.ProbeResult{
			Name:      n.Name,
			Location:  []float64{n.Lng, n.Lat},
			Reachable: reachable,
			LatencyMs: latency,
		})
	}

	return snap
}
