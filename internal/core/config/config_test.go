// internal/core/config/config_test.go
package config

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.DataDir == "" {
		t.Error("expected non-empty data dir")
	}
	if cfg.ScanInterval.String() != "5m0s" {
		t.Errorf("expected 5m scan interval, got %s", cfg.ScanInterval)
	}
	if cfg.DBDriver != "sqlite" {
		t.Errorf("expected sqlite driver, got %s", cfg.DBDriver)
	}
}

func TestConfigFromFlags(t *testing.T) {
	cfg := Default()
	args := []string{"--port", "9090", "--scan-interval", "10m"}
	err := cfg.ParseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.ScanInterval.String() != "10m0s" {
		t.Errorf("expected 10m scan interval, got %s", cfg.ScanInterval)
	}
}
