package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/netmap/netmap/internal/api/handlers"
	"github.com/netmap/netmap/internal/api/ws"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
	"github.com/netmap/netmap/internal/store/sqlite"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	f, _ := os.CreateTemp("", "netmap-integration-*.db")
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	db, err := sqlite.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	s := &store.Store{
		Devices:  sqlite.NewDeviceRepo(db),
		Networks: sqlite.NewNetworkRepo(db),
		Scans:    sqlite.NewScanRepo(db),
	}

	hub := ws.NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.Stop() })

	scanHandler := handlers.NewScanHandler(s.Scans)
	router := NewRouter(s, hub, scanHandler)
	return httptest.NewServer(router)
}

func TestDeviceCRUD(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Create
	body := `{"hostname":"test-server","ip_addresses":["10.0.0.1"]}`
	resp, err := http.Post(srv.URL+"/api/v1/devices", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", resp.StatusCode)
	}
	var created models.Device
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Hostname != "test-server" {
		t.Errorf("expected test-server, got %s", created.Hostname)
	}

	// List
	resp, _ = http.Get(srv.URL + "/api/v1/devices")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", resp.StatusCode)
	}
	var list models.ListResult[models.Device]
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()

	if list.Total != 1 {
		t.Errorf("expected 1 device, got %d", list.Total)
	}

	// Get
	resp, _ = http.Get(srv.URL + "/api/v1/devices/" + created.ID)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Delete
	req, _ := http.NewRequest("DELETE", srv.URL+"/api/v1/devices/"+created.ID, nil)
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deleted
	resp, _ = http.Get(srv.URL + "/api/v1/devices/" + created.ID)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestSystemStatus(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/system/status")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var status map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&status)
	resp.Body.Close()

	if _, ok := status["version"]; !ok {
		t.Error("expected version in status")
	}
}
