package persistence

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/device"
	"m2cp/tools"
	"os"
	"path"
)

var (
	pathSnapCommon  string
	pathSnapCurrent string
)

// PathSnapCommon returns a path to the writable directory across this snaps revisions - it always works, on a real device just as in an IDE with virtual-device
func PathSnapCommon() (string, error) {
	var err error

	if pathSnapCommon != "" {
		return pathSnapCommon, nil
	}

	p := os.Getenv("SNAP_COMMON")
	if p != "" {
		p, err = tools.PathAbs(p)
		if err != nil {
			return "", fmt.Errorf("cannot get path: %s", err)
		}
		pathSnapCommon = p
		return p, nil
	}

	virtualDevice := os.Getenv("M2CP_VIRTUAL_DEVICE")
	if virtualDevice != "" {
		return pathVirtualDeviceWriteable(contextplus.NewContextPlus(), "common")
	}

	snapName := os.Getenv("SNAP")
	if snapName != "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot aquire path of user home dir")
		}
		if p, err = tools.PathAbs(path.Join(home, ".m2cp", "snaps", snapName, "common")); err != nil {
			return "", fmt.Errorf("cannot get path: %s", err)
		}
		if err = os.MkdirAll(p, 0755); err != nil {
			return "", fmt.Errorf("cannot create directory: %s", err)
		}
		pathSnapCommon = p
		return p, nil
	}

	p, err = os.MkdirTemp("", "m2cp")
	if err != nil {
		return "", fmt.Errorf("cannot create temporary directory: %s", err)
	}
	pathSnapCommon = p
	return p, nil
}

// PathSnapCurrent returns a path to the writable directory for the current snap revision - it always works, on a real device just as in an IDE with virtual-device
func PathSnapCurrent() (string, error) {
	var err error

	if pathSnapCurrent != "" {
		return pathSnapCurrent, nil
	}

	p := os.Getenv("SNAP_DATA")
	if p != "" {
		p, err = tools.PathAbs(p)
		if err != nil {
			return "", fmt.Errorf("cannot get path: %s", err)
		}
		pathSnapCurrent = p
		return p, nil
	}

	virtualDevice := os.Getenv("M2CP_VIRTUAL_DEVICE")
	if virtualDevice != "" {
		return pathVirtualDeviceWriteable(contextplus.NewContextPlus(), "current")
	}

	snapName := os.Getenv("SNAP")
	if snapName != "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot aquire path of user home dir")
		}
		if p, err = tools.PathAbs(path.Join(home, ".m2cp", "snaps", snapName, "current")); err != nil {
			return "", fmt.Errorf("cannot get path: %s", err)
		}
		if err = os.MkdirAll(p, 0755); err != nil {
			return "", fmt.Errorf("cannot create directory: %s", err)
		}
		pathSnapCurrent = p
		return p, nil
	}

	p, err = os.MkdirTemp("", "m2cp")
	if err != nil {
		return "", fmt.Errorf("cannot create temporary directory: %s", err)
	}
	pathSnapCurrent = p
	return p, nil
}

func pathVirtualDeviceWriteable(ctp m2cp.ContextPlus, commonCurrent string) (p string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot aquire path of user home dir")
	}
	deviceSerial, err := device.GetName(contextplus.NewContextPlus())
	if err != nil {
		return "", fmt.Errorf("cannot get device serial: %s", err)
	}
	snapName := os.Getenv("SNAP")
	if snapName == "" {
		return "", fmt.Errorf("SNAP environment variable is required when working with connection to virtual device")
	}
	p, err = tools.PathAbs(path.Join(home, ".m2cp", "virtual-devices", deviceSerial, "var", "snap", snapName, "common"))
	if err != nil {
		return "", fmt.Errorf("cannot get path: %s", err)
	}

	if err = os.MkdirAll(p, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory: %s", err)
	}
	return p, nil
}
