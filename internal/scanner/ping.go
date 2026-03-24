package scanner

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

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

			alive := false
			latency := float64(0)

			start := time.Now()
			conn, err := net.DialTimeout("tcp", ip+":80", p.timeout)
			if err == nil {
				conn.Close()
				alive = true
				latency = float64(time.Since(start).Microseconds()) / 1000.0
			}

			if !alive {
				for _, port := range []string{":443", ":22", ":445"} {
					start = time.Now()
					conn, err := net.DialTimeout("tcp", ip+port, p.timeout)
					if err == nil {
						conn.Close()
						alive = true
						latency = float64(time.Since(start).Microseconds()) / 1000.0
						break
					}
				}
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
