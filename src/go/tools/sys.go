package tools

import (
	"runtime"
	"time"
)

func GetArch() string {
	switch runtime.GOOS {
	case "windows":
		return "windows-amd64"
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-amd64"
		case "arm64":
			return "linux-arm64"
		default:
			return "linux-unsupported"
		}
	default:
		return "unsupported"
	}
}

// TimeNow returns time in the precision used by the db
func TimeNow() time.Time {
	return time.Now().UTC().Round(time.Duration(10) * time.Millisecond)
}
