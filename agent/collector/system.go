package collector

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type SystemMetrics struct {
	CPUPercent  float64 `json:"cpuPercent"`
	MemPercent  float64 `json:"memPercent"`
	DiskPercent float64 `json:"diskPercent"`
}

func CollectSystem() SystemMetrics {
	var m SystemMetrics

	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		m.CPUPercent = round(p[0], 1)
	}
	if v, err := mem.VirtualMemory(); err == nil {
		m.MemPercent = round(v.UsedPercent, 1)
	}
	if d, err := disk.Usage("/"); err == nil {
		m.DiskPercent = round(d.UsedPercent, 1)
	}
	return m
}

func round(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int64(v*pow+0.5)) / pow
}
