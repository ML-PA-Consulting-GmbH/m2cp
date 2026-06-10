package tools

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Forwarded-For")
	//IPAddress := r.Header.Get("X-Real-Ip")
	//if IPAddress == "" {
	//}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return strings.Split(IPAddress, ":")[0]
}

func GetUsedPortsOs() ([]int, error) {
	var usedPorts []int
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("netstat", "-an", "-p", "tcp")
	} else {
		cmd = exec.Command("/usr/bin/bash", "-c", "ss -tlnH | awk '{print $4}' | grep -Eo ':[0-9]*' | grep -Eo '[0-9]*' | sort -u")
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	retrievedPorts := strings.Split(string(out), "\n")
	if len(retrievedPorts) == 0 {
		return usedPorts, nil
	}

	for _, line := range retrievedPorts {
		if len(line) == 0 {
			continue
		}
		if runtime.GOOS == "windows" {
			// On Windows, extract the port from the "Local Address" field
			fields := strings.Fields(line)
			if len(fields) < 2 || !strings.Contains(fields[1], ":") {
				continue
			}
			addressParts := strings.Split(fields[1], ":")
			port := addressParts[len(addressParts)-1]
			p, err := strconv.Atoi(port)
			if err != nil {
				continue
			}
			usedPorts = append(usedPorts, p)
		} else {
			// On Linux, the port is already extracted
			p, err := strconv.Atoi(line)
			if err != nil {
				continue
			}
			usedPorts = append(usedPorts, p)
		}
	}
	return usedPorts, nil
}

// IsUdpPortInUse Check, if the given UDP port is already in use by any process in the system
func IsUdpPortInUse(port string) bool {
	address := fmt.Sprintf(":%s", port)
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		return true
	}
	defer conn.Close()
	return false
}
