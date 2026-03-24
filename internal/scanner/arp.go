package scanner

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/mdlayher/arp"
	"github.com/netmap/netmap/internal/core/models"
)

type NetworkProber struct {
	timeout time.Duration
	workers int
}

func NewNetworkProber(timeout time.Duration, workers int) *NetworkProber {
	return &NetworkProber{timeout: timeout, workers: workers}
}

// ARPScan attempts a real ARP sweep of the subnet.
// If raw socket access is unavailable (not root / no CAP_NET_RAW),
// it falls back to enumerating all IPs for the TCP probe stage.
func (p *NetworkProber) ARPScan(ctx context.Context, subnet string) ([]models.HostResult, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("parse subnet: %w", err)
	}

	// Try real ARP first.
	if results, err := p.arpSweep(ctx, ipnet); err == nil {
		return results, nil
	}

	// Fall back: return all IPs in the subnet for the TCP probe stage.
	return enumerateSubnet(ctx, ipnet), nil
}

// arpSweep performs a real ARP sweep using raw sockets.
func (p *NetworkProber) arpSweep(ctx context.Context, ipnet *net.IPNet) ([]models.HostResult, error) {
	iface, err := interfaceForSubnet(ipnet)
	if err != nil {
		return nil, err
	}

	client, err := arp.Dial(iface)
	if err != nil {
		return nil, fmt.Errorf("arp dial: %w", err)
	}
	defer client.Close()

	// Collect ARP replies in a goroutine.
	var mu sync.Mutex
	replies := make(map[string]net.HardwareAddr)
	done := make(chan struct{})

	go func() {
		defer close(done)
		deadline := time.Now().Add(3 * time.Second)
		_ = client.SetReadDeadline(deadline)
		for {
			pkt, _, err := client.Read()
			if err != nil {
				return
			}
			if pkt.Operation == arp.OperationReply {
				mu.Lock()
				replies[pkt.SenderIP.String()] = pkt.SenderHardwareAddr
				mu.Unlock()
			}
		}
	}()

	// Send ARP requests to every host in the subnet.
	ip := cloneIP(ipnet.IP.Mask(ipnet.Mask))
	incIP(ip)
	broadcast := lastIP(ipnet)
	for ; ipnet.Contains(ip) && !ip.Equal(broadcast); incIP(ip) {
		if ctx.Err() != nil {
			break
		}
		target := cloneIP(ip)
		if addr, ok := netip.AddrFromSlice(target); ok {
			_ = client.Request(addr.Unmap())
		}
		time.Sleep(2 * time.Millisecond) // gentle pacing
	}

	// Wait for replies to arrive (up to the deadline).
	timer := time.NewTimer(3 * time.Second)
	select {
	case <-done:
	case <-timer.C:
	}
	timer.Stop()

	mu.Lock()
	defer mu.Unlock()

	var results []models.HostResult
	for ipStr, mac := range replies {
		hostname, _ := resolveHostname(ipStr)
		results = append(results, models.HostResult{
			IP:       ipStr,
			MAC:      mac.String(),
			Hostname: hostname,
			Status:   models.HostStatus("up"),
		})
	}
	return results, nil
}

// enumerateSubnet returns a HostResult skeleton for every IP in the subnet.
func enumerateSubnet(ctx context.Context, ipnet *net.IPNet) []models.HostResult {
	broadcast := lastIP(ipnet)
	ip := cloneIP(ipnet.IP.Mask(ipnet.Mask))
	incIP(ip)

	var hosts []models.HostResult
	for ; ipnet.Contains(ip); incIP(ip) {
		if ctx.Err() != nil {
			break
		}
		if ip.Equal(broadcast) {
			break
		}
		hosts = append(hosts, models.HostResult{
			IP:     ip.String(),
			Status: models.HostStatus("unknown"),
		})
	}
	return hosts
}

// interfaceForSubnet finds the network interface whose address is in ipnet.
func interfaceForSubnet(ipnet *net.IPNet) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ipnet.Contains(ip) {
				return &iface, nil
			}
		}
	}
	return nil, fmt.Errorf("no interface found for subnet %s", ipnet)
}

// resolveHostname does a reverse DNS lookup, returning empty string on failure.
func resolveHostname(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return "", err
	}
	// Strip trailing dot from PTR records.
	name := names[0]
	if len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}
	return name, nil
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

func lastIP(ipnet *net.IPNet) net.IP {
	ip := cloneIP(ipnet.IP.Mask(ipnet.Mask))
	for i := range ip {
		ip[i] |= ^ipnet.Mask[i]
	}
	return ip
}

func countHosts(subnet string) int {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return 0
	}
	ones, bits := ipnet.Mask.Size()
	return (1 << (bits - ones)) - 2
}
