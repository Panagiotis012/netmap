// internal/store/sqlite/scan_repo_test.go
package sqlite

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

func TestScanCreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewScanRepo(db)
	ctx := context.Background()

	now := time.Now()
	scan := &models.ScanJob{
		ID:        "scan-1",
		Type:      models.ScanDiscovery,
		Target:    "192.168.1.0/24",
		Status:    models.ScanCompleted,
		StartedAt: &now,
		Results:   json.RawMessage(`{"hosts":[],"stats":{"hosts_scanned":254}}`),
	}

	if err := repo.Create(ctx, scan); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(ctx, "scan-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != models.ScanDiscovery {
		t.Errorf("expected discovery, got %s", got.Type)
	}
	if got.Results == nil {
		t.Error("expected results to be set")
	}
}

func TestScanPendingNullStartedAt(t *testing.T) {
	db := testDB(t)
	repo := NewScanRepo(db)
	ctx := context.Background()

	// Pending scan has nil StartedAt
	scan := &models.ScanJob{
		ID:        "pending-1",
		Type:      models.ScanDiscovery,
		Target:    "10.0.0.0/24",
		Status:    models.ScanPending,
		StartedAt: nil,
		Results:   nil,
	}
	if err := repo.Create(ctx, scan); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(ctx, "pending-1")
	if err != nil {
		t.Fatalf("expected no error for pending scan: %v", err)
	}
	if got.StartedAt != nil {
		t.Errorf("expected nil StartedAt for pending scan, got %v", got.StartedAt)
	}
}
