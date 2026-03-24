package scanner

import (
	"bytes"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

// DefaultGateway returns the default gateway IP for the system, or empty string.
func DefaultGateway() string {
	switch runtime.GOOS {
	case "darwin":
		return gatewayDarwin()
	case "linux":
		return gatewayLinux()
	default:
		return ""
	}
}

func gatewayDarwin() string {
	out, err := exec.Command("route", "-n", "get", "default").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ip := net.ParseIP(parts[1])
				if ip != nil {
					return ip.String()
				}
			}
		}
	}
	return ""
}

func gatewayLinux() string {
	out, err := exec.Command("ip", "route", "show", "default").Output()
	if err != nil {
		// Fallback to route command
		out, err = exec.Command("route", "-n").Output()
		if err != nil {
			return ""
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] == "0.0.0.0" {
				if net.ParseIP(fields[1]) != nil {
					return fields[1]
				}
			}
		}
		return ""
	}
	// "default via 192.168.1.1 dev eth0"
	fields := strings.Fields(string(bytes.TrimSpace(out)))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			if net.ParseIP(fields[i+1]) != nil {
				return fields[i+1]
			}
		}
	}
	return ""
}
