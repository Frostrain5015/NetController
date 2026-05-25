package collector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

type pm2Process struct {
	Name  string `json:"name"`
	PID   int    `json:"pid"`
	Pm2Env struct {
		Status string `json:"status"`
	} `json:"pm2_env"`
	Monit struct {
		CPU    float64 `json:"cpu"`
		Memory int64   `json:"memory"`
	} `json:"monit"`
}

type pm2Info struct {
	Alive bool
	PID   int32
	CPU   float64
	MemMB uint64
}

func getPM2Map() map[string]pm2Info {
	out, err := exec.Command("pm2", "jlist").Output()
	if err != nil {
		return nil
	}
	var processes []pm2Process
	if err := json.Unmarshal(out, &processes); err != nil {
		return nil
	}
	m := make(map[string]pm2Info)
	for _, p := range processes {
		m[p.Name] = pm2Info{
			Alive: p.Pm2Env.Status == "online",
			PID:   int32(p.PID),
			CPU:   p.Monit.CPU,
			MemMB: uint64(p.Monit.Memory) / (1024 * 1024),
		}
	}
	return m
}

func FindProcessByName(name string) (pid int32, alive bool) {
	// 先查 PM2
	if pm2Map := getPM2Map(); pm2Map != nil {
		if info, ok := pm2Map[name]; ok {
			return info.PID, info.Alive
		}
	}
	// 兜底：ps aux 搜索
	procs, err := process.Processes()
	if err != nil {
		return 0, false
	}
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}
		if strings.Contains(n, name) {
			return p.Pid, true
		}
		// 也检查命令行参数（对 java 进程等有用）
		cmdline, err := p.Cmdline()
		if err != nil {
			continue
		}
		if strings.Contains(cmdline, name) {
			return p.Pid, true
		}
	}
	return 0, false
}

func CheckPort(port int) bool {
	if port <= 0 {
		return false
	}
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		out, err = exec.Command("netstat", "-tlnp").Output()
		if err != nil {
			return false
		}
	}
	target := fmt.Sprintf(":%d ", port)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, target) {
			return true
		}
	}
	return false
}

func GetProcessCPUAndMem(name string) (cpu float64, memMB uint64) {
	if pm2Map := getPM2Map(); pm2Map != nil {
		if info, ok := pm2Map[name]; ok && info.Alive {
			return info.CPU, info.MemMB
		}
	}
	// 兜底：gopsutil
	procs, err := process.Processes()
	if err != nil {
		return 0, 0
	}
	for _, p := range procs {
		n, err := p.Name()
		if err != nil {
			continue
		}
		if strings.Contains(n, name) {
			cpuPct, _ := p.CPUPercent()
			memInfo, _ := p.MemoryInfo()
			if memInfo != nil {
				return cpuPct, memInfo.RSS / (1024 * 1024)
			}
			return cpuPct, 0
		}
	}
	return 0, 0
}
