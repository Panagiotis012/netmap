package scanner

import (
	"context"
	"net"
	"strconv"

	"github.com/netmap/netmap/internal/core/models"
)

var serviceMap = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
	80: "http", 110: "pop3", 143: "imap", 443: "https", 445: "smb",
	3306: "mysql", 3389: "rdp", 5432: "postgres", 5900: "vnc",
	8080: "http-alt", 8443: "https-alt",
}

func (p *NetworkProber) PortScan(ctx context.Context, ip string, ports []int) ([]models.PortResult, error) {
	var results []models.PortResult
	for _, port := range ports {
		if ctx.Err() != nil {
			break
		}
		addr := net.JoinHostPort(ip, strconv.Itoa(port))
		conn, err := net.DialTimeout("tcp", addr, p.timeout)
		if err == nil {
			conn.Close()
			service := serviceMap[port]
			if service == "" {
				service = "unknown"
			}
			results = append(results, models.PortResult{
				Number: port, Protocol: "tcp", Service: service, State: "open",
			})
		}
	}
	return results, nil
}
