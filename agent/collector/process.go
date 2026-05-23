package collector

import (
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func FindProcessByName(name string) (pid int32, alive bool) {
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
	}
	return 0, false
}

func CheckPort(port int) bool {
	// Use ss/netstat approach — try connecting to localhost:port
	// For Linux, use ss for speed; fallback to exec
	out, err := exec.Command("ss", "-tlnp").Output()
	if err != nil {
		// fallback: try netstat
		out, err = exec.Command("netstat", "-tlnp").Output()
		if err != nil {
			return false
		}
	}
	target := ":"
	if port > 0 {
		target = ":"
		// search for :{port} in output
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, target) {
				return true
			}
		}
	}
	return false
}
