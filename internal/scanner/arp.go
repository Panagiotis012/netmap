package scanner

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

type NetworkProber struct {
	timeout time.Duration
}

func NewNetworkProber(timeout time.Duration) *NetworkProber {
	return &NetworkProber{timeout: timeout}
}

func (p *NetworkProber) ARPScan(ctx context.Context, subnet string) ([]models.HostResult, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse subnet: %w", err)
	}

	// Phase 1: TCP-based discovery (ARP requires raw sockets / root)
	var hosts []models.HostResult
	for ip := cloneIP(ipnet.IP.Mask(ipnet.Mask)); ipnet.Contains(ip); incIP(ip) {
		if ctx.Err() != nil {
			break
		}
		hosts = append(hosts, models.HostResult{
			IP:     ip.String(),
			Status: models.HostStatus("unknown"),
		})
	}
	return hosts, nil
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func countHosts(subnet string) int {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return 0
	}
	ones, bits := ipnet.Mask.Size()
	return (1 << (bits - ones)) - 2
}
