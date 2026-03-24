package scanner

import (
	"context"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// probePorts are checked in order; first success marks the host as alive.
var probePorts = []string{
	":80", ":443", ":22", ":445",
	":8080", ":8443", ":23", ":21",
	":53", ":3389", ":5900", ":7547",
}

func (p *NetworkProber) PingSweep(ctx context.Context, hosts []string, progress ProgressFunc) ([]PingResult, error) {
	var mu sync.Mutex
	var results []PingResult
	var wg sync.WaitGroup
	var scanned int32

	workers := p.workers
	if workers <= 0 {
		workers = 50
	}
	sem := make(chan struct{}, workers)
	total := len(hosts)

	for _, host := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			alive, latency := tcpProbe(ip, p.timeout)
			if !alive {
				alive, latency = icmpPing(ip)
			}

			if alive {
				hostname, _ := resolveHostname(ip)
				mu.Lock()
				results = append(results, PingResult{IP: ip, Alive: true, LatencyMs: latency, Hostname: hostname})
				mu.Unlock()
			}

			n := int(atomic.AddInt32(&scanned, 1))
			if progress != nil {
				mu.Lock()
				found := len(results)
				mu.Unlock()
				progress(n, total, found)
			}
		}(host)
	}

	wg.Wait()
	return results, nil
}

// tcpProbe tries common TCP ports to determine if a host is reachable.
func tcpProbe(ip string, timeout time.Duration) (alive bool, latencyMs float64) {
	for _, port := range probePorts {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", ip+port, timeout)
		if err == nil {
			conn.Close()
			return true, float64(time.Since(start).Microseconds()) / 1000.0
		}
	}
	return false, 0
}

// icmpPing uses the OS ping command as an unprivileged ICMP fallback.
func icmpPing(ip string) (alive bool, latencyMs float64) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("ping", "-c", "1", "-W", "1000", "-t", "2", ip)
	case "linux":
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	default:
		return false, 0
	}

	start := time.Now()
	if err := cmd.Run(); err != nil {
		return false, 0
	}
	return true, float64(time.Since(start).Microseconds()) / 1000.0
}
