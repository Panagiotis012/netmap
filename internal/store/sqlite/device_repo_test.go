// internal/store/sqlite/device_repo_test.go
package sqlite

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	f, err := os.CreateTemp("", "netmap-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDeviceCreateAndGet(t *testing.T) {
	db := testDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	device := &models.Device{
		ID:              "test-id-1",
		Hostname:        "test-host",
		IPAddresses:     []string{"192.168.1.10"},
		MACAddresses:    []string{"aa:bb:cc:dd:ee:ff"},
		Status:          models.StatusOnline,
		DiscoveryMethod: models.DiscoveryScan,
		FirstSeenAt:     time.Now(),
		LastSeenAt:      time.Now(),
		Tags:            []string{"server", "linux"},
	}

	err := repo.Create(ctx, device)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByID(ctx, "test-id-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Hostname != "test-host" {
		t.Errorf("expected hostname test-host, got %s", got.Hostname)
	}
	if len(got.IPAddresses) != 1 || got.IPAddresses[0] != "192.168.1.10" {
		t.Errorf("unexpected IPs: %v", got.IPAddresses)
	}
	if len(got.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(got.Tags))
	}
}

func TestDeviceList(t *testing.T) {
	db := testDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		d := &models.Device{
			ID:          fmt.Sprintf("dev-%d", i),
			Hostname:    fmt.Sprintf("host-%d", i),
			IPAddresses: []string{fmt.Sprintf("192.168.1.%d", i+10)},
			Status:      models.StatusOnline,
			FirstSeenAt: time.Now(),
			LastSeenAt:  time.Now(),
			Tags:        []string{},
		}
		if err := repo.Create(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	result, err := repo.List(ctx, models.ListParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 total, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestDeviceGetByMAC(t *testing.T) {
	db := testDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	device := &models.Device{
		ID:           "mac-test",
		Hostname:     "mac-host",
		MACAddresses: []string{"11:22:33:44:55:66"},
		Status:       models.StatusOnline,
		FirstSeenAt:  time.Now(),
		LastSeenAt:   time.Now(),
		Tags:         []string{},
	}
	if err := repo.Create(ctx, device); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByMAC(ctx, "11:22:33:44:55:66")
	if err != nil {
		t.Fatalf("get by mac: %v", err)
	}
	if got.ID != "mac-test" {
		t.Errorf("expected mac-test, got %s", got.ID)
	}
}

func TestDeviceCountByStatus(t *testing.T) {
	db := testDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	statuses := []models.DeviceStatus{models.StatusOnline, models.StatusOnline, models.StatusOffline}
	for i, s := range statuses {
		d := &models.Device{
			ID: fmt.Sprintf("count-%d", i), Status: s,
			FirstSeenAt: time.Now(), LastSeenAt: time.Now(),
			Tags: []string{},
		}
		if err := repo.Create(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	on, off, unk, err := repo.CountByStatus(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if on != 2 || off != 1 || unk != 0 {
		t.Errorf("expected 2/1/0, got %d/%d/%d", on, off, unk)
	}
}

func TestDeviceErrNotFound(t *testing.T) {
	db := testDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected store.ErrNotFound, got %v", err)
	}
}
