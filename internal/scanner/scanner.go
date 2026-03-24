package scanner

import (
	"context"
	"sync"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

type PingResult struct {
	IP        string
	Alive     bool
	LatencyMs float64
	Hostname  string
}

// ProgressFunc is called after each host is probed during PingSweep.
// scanned = hosts attempted so far, total = total hosts, found = alive hosts found so far.
type ProgressFunc func(scanned, total, found int)

type Prober interface {
	ARPScan(ctx context.Context, subnet string) ([]models.HostResult, error)
	PingSweep(ctx context.Context, hosts []string, progress ProgressFunc) ([]PingResult, error)
	PortScan(ctx context.Context, ip string, ports []int) ([]models.PortResult, error)
}

var CommonPorts = []int{
	21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995,
	1723, 3306, 3389, 5432, 5900, 8080, 8443, 8888,
}

type Scanner struct {
	prober     Prober
	workers    int
	portRanges []int
}

func NewScanner(prober Prober, workers int, portRanges []int) *Scanner {
	if workers <= 0 {
		workers = 50
	}
	if len(portRanges) == 0 {
		portRanges = CommonPorts
	}
	return &Scanner{prober: prober, workers: workers, portRanges: portRanges}
}

func (s *Scanner) Scan(ctx context.Context, subnet string, mode models.ScanType, progress ProgressFunc) (*models.ScanResults, error) {
	start := time.Now()

	// Step 1: ARP/discovery
	hosts, err := s.prober.ARPScan(ctx, subnet)
	if err != nil {
		return nil, err
	}

	// Step 2: Ping sweep for latency + confirm unverified hosts.
	// Hosts with a MAC came from real ARP — already confirmed alive.
	// Hosts without a MAC came from subnet enumeration — need TCP confirmation.
	var arpConfirmed []models.HostResult
	var needsProbe []string
	for _, h := range hosts {
		if h.MAC != "" {
			arpConfirmed = append(arpConfirmed, h)
		} else {
			needsProbe = append(needsProbe, h.IP)
		}
	}

	pingResults, err := s.prober.PingSweep(ctx, append(needsProbe, func() []string {
		ips := make([]string, len(arpConfirmed))
		for i, h := range arpConfirmed {
			ips[i] = h.IP
		}
		return ips
	}()...), progress)
	if err != nil {
		return nil, err
	}

	pingMap := make(map[string]PingResult)
	for _, p := range pingResults {
		pingMap[p.IP] = p
	}

	// ARP-confirmed hosts are always alive; enrich with latency/hostname if ping responded.
	for i, h := range arpConfirmed {
		if p, ok := pingMap[h.IP]; ok {
			arpConfirmed[i].LatencyMs = p.LatencyMs
			if h.Hostname == "" && p.Hostname != "" {
				arpConfirmed[i].Hostname = p.Hostname
			}
		}
	}

	// Enumerated hosts only survive if TCP probe confirmed them.
	var tcpAlive []models.HostResult
	for _, ip := range needsProbe {
		if p, ok := pingMap[ip]; ok {
			tcpAlive = append(tcpAlive, models.HostResult{
				IP:        ip,
				Hostname:  p.Hostname,
				LatencyMs: p.LatencyMs,
				Status:    models.HostStatus("up"),
			})
		}
	}

	hosts = append(arpConfirmed, tcpAlive...)

	// Step 3: Port scan (port and full modes)
	if mode == models.ScanPort || mode == models.ScanFull {
		s.portScanHosts(ctx, hosts)
	}

	elapsed := time.Since(start)
	return &models.ScanResults{
		Hosts: hosts,
		Stats: models.ScanStats{
			HostsScanned: countHosts(subnet),
			HostsUp:      len(hosts),
			DurationMs:   elapsed.Milliseconds(),
		},
	}, nil
}

func (s *Scanner) portScanHosts(ctx context.Context, hosts []models.HostResult) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.workers)

	for i := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			ports, err := s.prober.PortScan(ctx, hosts[idx].IP, s.portRanges)
			if err == nil {
				hosts[idx].Ports = ports
			}
		}(i)
	}
	wg.Wait()
}
