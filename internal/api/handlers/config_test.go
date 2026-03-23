package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netmap/netmap/internal/api/handlers"
	"github.com/netmap/netmap/internal/store/sqlite"
)

func newConfigHandlerForTest(t *testing.T) *handlers.ConfigHandler {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return handlers.NewConfigHandler(sqlite.NewConfigRepo(db))
}

func TestConfigGet_ReturnsDefaults(t *testing.T) {
	h := newConfigHandlerForTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/config", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["scan_interval"] != "5m" {
		t.Errorf("expected default scan_interval=5m, got %q", body["scan_interval"])
	}
	if body["scan_workers"] != "50" {
		t.Errorf("expected default scan_workers=50, got %q", body["scan_workers"])
	}
	if body["port_ranges"] != "22,80,443,8080,8443" {
		t.Errorf("unexpected port_ranges: %q", body["port_ranges"])
	}
}

func TestConfigPut_MergesKeys(t *testing.T) {
	h := newConfigHandlerForTest(t)

	body := bytes.NewBufferString(`{"scan_interval":"15m"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/system/config", body)
	w := httptest.NewRecorder()
	h.Put(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["scan_interval"] != "15m" {
		t.Errorf("expected updated scan_interval=15m, got %q", resp["scan_interval"])
	}
	// Other keys should still have defaults
	if resp["scan_workers"] != "50" {
		t.Errorf("expected default scan_workers=50, got %q", resp["scan_workers"])
	}
}

func TestConfigPut_RejectsUnknownKey(t *testing.T) {
	h := newConfigHandlerForTest(t)

	body := bytes.NewBufferString(`{"unknown_key":"value"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/system/config", body)
	w := httptest.NewRecorder()
	h.Put(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
