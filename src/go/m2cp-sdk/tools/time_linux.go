//go:build linux
// +build linux

package tools

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func timeSystemRunningOS() time.Duration {
	// For Linux systems
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if uptime, err := strconv.ParseFloat(parts[0], 64); err == nil {
				return time.Duration(uptime * float64(time.Second))
			}
		}
	}

	// Fallback to syscall
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err == nil {
		return time.Duration(info.Uptime) * time.Second
	}

	panic("Could not determine system uptime")
}
