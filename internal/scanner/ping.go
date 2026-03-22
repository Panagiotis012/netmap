package scanner

import (
	"context"
	"net"
	"sync"
	"time"
)

func (p *NetworkProber) PingSweep(ctx context.Context, hosts []string) ([]PingResult, error) {
	var mu sync.Mutex
	var results []PingResult
	var wg sync.WaitGroup

	workers := p.workers
	if workers <= 0 {
		workers = 50
	}
	sem := make(chan struct{}, workers)

	for _, host := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			alive := false
			latency := float64(0)

			// Try port 80 first, measuring only the successful connection
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
				mu.Lock()
				results = append(results, PingResult{IP: ip, Alive: true, LatencyMs: latency})
				mu.Unlock()
			}
		}(host)
	}

	wg.Wait()
	return results, nil
}
