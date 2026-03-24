package models_test

import (
	"encoding/json"
	"testing"

	"github.com/netmap/netmap/internal/core/models"
)

func TestScanCancelledConstant(t *testing.T) {
	if string(models.ScanCancelled) != "cancelled" {
		t.Errorf("expected ScanCancelled == 'cancelled', got %q", models.ScanCancelled)
	}
}

func TestScanProgressPayloadJSON(t *testing.T) {
	p := models.ScanProgressPayload{
		ScanID:       "abc",
		HostsScanned: 10,
		HostsTotal:   100,
		HostsFound:   3,
		Percent:      10,
		EtaSeconds:   45,
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	if out["scan_id"] != "abc" {
		t.Errorf("expected scan_id=abc, got %v", out["scan_id"])
	}
	if out["percent"].(float64) != 10 {
		t.Errorf("expected percent=10, got %v", out["percent"])
	}
}
