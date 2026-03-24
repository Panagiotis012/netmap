package sqlite_test

import (
	"context"
	"testing"

	"github.com/netmap/netmap/internal/store/sqlite"
)

func TestConfigRepo(t *testing.T) {
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := sqlite.NewConfigRepo(db)

	t.Run("Get returns empty string for unknown key", func(t *testing.T) {
		val := repo.Get(context.Background(), "missing_key")
		if val != "" {
			t.Errorf("expected empty, got %q", val)
		}
	})

	t.Run("Set and Get roundtrip", func(t *testing.T) {
		if err := repo.Set(context.Background(), "scan_interval", "15m"); err != nil {
			t.Fatal(err)
		}
		val := repo.Get(context.Background(), "scan_interval")
		if val != "15m" {
			t.Errorf("expected 15m, got %q", val)
		}
	})

	t.Run("Set overwrites existing key", func(t *testing.T) {
		repo.Set(context.Background(), "scan_workers", "50")
		repo.Set(context.Background(), "scan_workers", "100")
		val := repo.Get(context.Background(), "scan_workers")
		if val != "100" {
			t.Errorf("expected 100, got %q", val)
		}
	})

	t.Run("GetAll returns all stored keys", func(t *testing.T) {
		repo.Set(context.Background(), "k1", "v1")
		repo.Set(context.Background(), "k2", "v2")
		all := repo.GetAll(context.Background())
		if all["k1"] != "v1" || all["k2"] != "v2" {
			t.Errorf("unexpected GetAll result: %v", all)
		}
	})
}
