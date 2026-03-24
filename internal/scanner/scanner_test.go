package scanner

import (
	"context"
	"testing"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

type MockProber struct {
	hosts []models.HostResult
}

func (m *MockProber) ARPScan(ctx context.Context, subnet string) ([]models.HostResult, error) {
	return m.hosts, nil
}

func (m *MockProber) PingSweep(ctx context.Context, hosts []string, progress ProgressFunc) ([]PingResult, error) {
	var results []PingResult
	for _, h := range hosts {
		results = append(results, PingResult{IP: h, Alive: true, LatencyMs: 1.5})
	}
	return results, nil
}

func (m *MockProber) PortScan(ctx context.Context, ip string, ports []int) ([]models.PortResult, error) {
	return []models.PortResult{
		{Number: 22, Protocol: "tcp", Service: "ssh", State: "open"},
	}, nil
}

func TestScannerDiscoveryMode(t *testing.T) {
	mock := &MockProber{
		hosts: []models.HostResult{
			{IP: "192.168.1.10", MAC: "aa:bb:cc:dd:ee:ff", Status: models.HostUp},
			{IP: "192.168.1.11", MAC: "11:22:33:44:55:66", Status: models.HostUp},
		},
	}

	s := NewScanner(mock, 10, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := s.Scan(ctx, "192.168.1.0/24", models.ScanDiscovery, nil)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if results.Stats.HostsUp != 2 {
		t.Errorf("expected 2 hosts up, got %d", results.Stats.HostsUp)
	}
	if len(results.Hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(results.Hosts))
	}
}

func TestScannerPortMode(t *testing.T) {
	mock := &MockProber{
		hosts: []models.HostResult{
			{IP: "192.168.1.10", MAC: "aa:bb:cc:dd:ee:ff", Status: models.HostUp},
		},
	}

	s := NewScanner(mock, 10, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := s.Scan(ctx, "192.168.1.0/24", models.ScanPort, nil)
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}
	if len(results.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(results.Hosts))
	}
	if len(results.Hosts[0].Ports) == 0 {
		t.Error("expected ports to be scanned in port mode")
	}
}
