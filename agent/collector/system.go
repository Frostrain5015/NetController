package collector

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetrics struct {
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float64 `json:"memPercent"`
	MemUsedMB   uint64  `json:"memUsedMB"`
	MemTotalMB  uint64  `json:"memTotalMB"`
	DiskPercent float64 `json:"diskPercent"`
	DiskUsed    string  `json:"diskUsed"`
	DiskTotal   string  `json:"diskTotal"`
}

func CollectSystem() SystemMetrics {
	var m SystemMetrics

	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		m.CPUPercent = round(p[0], 1)
	}
	if v, err := mem.VirtualMemory(); err == nil {
		m.MemPercent = round(v.UsedPercent, 1)
		m.MemUsedMB = v.Used / (1024 * 1024)
		m.MemTotalMB = v.Total / (1024 * 1024)
	}
	if d, err := disk.Usage("/"); err == nil {
		m.DiskPercent = round(d.UsedPercent, 1)
		m.DiskUsed = formatBytes(d.Used)
		m.DiskTotal = formatBytes(d.Total)
	}
	return m
}

func formatBytes(b uint64) string {
	gb := float64(b) / (1024 * 1024 * 1024)
	if gb >= 1 {
		return fmt.Sprintf("%.1fG", gb)
	}
	mb := float64(b) / (1024 * 1024)
	return fmt.Sprintf("%.0fM", mb)
}

func round(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}
