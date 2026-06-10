//go:build windows
// +build windows

package tools

import (
	"syscall"
	"time"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	getTickCount64Proc = kernel32.NewProc("GetTickCount64")
)

func timeSystemRunningOS() time.Duration {
	ret, _, _ := getTickCount64Proc.Call()
	// GetTickCount64 returns the number of milliseconds since system boot
	return time.Duration(ret) * time.Millisecond
}
