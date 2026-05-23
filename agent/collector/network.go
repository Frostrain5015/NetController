package collector

import (
	"net"
	"time"
)

type ProbeResult struct {
	Name      string    `json:"name"`
	Location  []float64 `json:"location"`
	Reachable bool      `json:"reachable"`
	LatencyMs int64     `json:"latencyMs"`
}

func ProbeTCP(host string, timeout time.Duration) (reachable bool, latencyMs int64) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false, -1
	}
	latencyMs = time.Since(start).Milliseconds()
	conn.Close()
	return true, latencyMs
}
