//go:build linux
// +build linux

package tools

import (
	"context"
	"os"
	"runtime/trace"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func timeSystemRunningOS() time.Duration {
	// For Linux systems
	traceReadProc := trace.StartRegion(context.TODO(), "timeSystemRunningOS.readProcUptime")
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if uptime, err := strconv.ParseFloat(parts[0], 64); err == nil {
				traceReadProc.End()
				return time.Duration(uptime * float64(time.Second))
			}
		}
	}
	traceReadProc.End()

	// Fallback to syscall
	traceSyscall := trace.StartRegion(context.TODO(), "timeSystemRunningOS.syscallSysinfo")
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err == nil {
		traceSyscall.End()
		return time.Duration(info.Uptime) * time.Second
	}
	traceSyscall.End()

	panic("Could not determine system uptime")
}
