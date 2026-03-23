package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netmap/netmap/internal/api/handlers"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store/sqlite"
)

func newScanTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestScanTrigger_Returns202WithID(t *testing.T) {
	db := newScanTestDB(t)
	repo := sqlite.NewScanRepo(db)
	h := handlers.NewScanHandler(repo)

	triggered := make(chan struct{}, 1)
	h.ScanTrigger = func(ctx context.Context, id string, scanType models.ScanType, target string) {
		triggered <- struct{}{}
	}

	body := bytes.NewBufferString(`{"type":"discovery","target":"192.168.1.0/24"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body)
	w := httptest.NewRecorder()
	h.Trigger(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["id"] == "" {
		t.Error("expected non-empty id in response")
	}
	if resp["status"] != "running" {
		t.Errorf("expected status=running, got %q", resp["status"])
	}

	select {
	case <-triggered:
	case <-time.After(time.Second):
		t.Error("ScanTrigger not called")
	}
}

func TestScanTrigger_409WhenScanAlreadyRunning(t *testing.T) {
	db := newScanTestDB(t)
	repo := sqlite.NewScanRepo(db)
	h := handlers.NewScanHandler(repo)

	h.ScanTrigger = func(ctx context.Context, id string, scanType models.ScanType, target string) {
		time.Sleep(500 * time.Millisecond)
	}

	body1 := bytes.NewBufferString(`{"type":"discovery","target":"192.168.1.0/24"}`)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body1)
	w1 := httptest.NewRecorder()
	h.Trigger(w1, req1)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("first trigger: expected 202, got %d", w1.Code)
	}

	time.Sleep(10 * time.Millisecond)

	body2 := bytes.NewBufferString(`{"type":"discovery","target":"192.168.1.0/24"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body2)
	w2 := httptest.NewRecorder()
	h.Trigger(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("second trigger: expected 409, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestScanCancel_RunningReturns204AndGoroutineExits(t *testing.T) {
	db := newScanTestDB(t)
	repo := sqlite.NewScanRepo(db)
	h := handlers.NewScanHandler(repo)

	var capturedID string
	done := make(chan struct{})
	h.ScanTrigger = func(ctx context.Context, id string, scanType models.ScanType, target string) {
		capturedID = id
		<-ctx.Done()
		close(done)
	}

	body := bytes.NewBufferString(`{"type":"discovery","target":"192.168.1.0/24"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", body)
	w := httptest.NewRecorder()
	h.Trigger(w, req)

	time.Sleep(10 * time.Millisecond)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", capturedID)
	cancelReq := httptest.NewRequest(http.MethodDelete, "/api/v1/scans/"+capturedID, nil)
	cancelReq = cancelReq.WithContext(context.WithValue(cancelReq.Context(), chi.RouteCtxKey, rctx))
	cw := httptest.NewRecorder()
	h.Cancel(cw, cancelReq)

	if cw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", cw.Code, cw.Body.String())
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("scan goroutine did not exit after cancel")
	}
}

func TestScanCancel_UnknownIDReturns404(t *testing.T) {
	db := newScanTestDB(t)
	repo := sqlite.NewScanRepo(db)
	h := handlers.NewScanHandler(repo)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent-uuid")
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/scans/nonexistent-uuid", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	h.Cancel(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestScanCancel_CompletedScanReturns409(t *testing.T) {
	db := newScanTestDB(t)
	repo := sqlite.NewScanRepo(db)
	h := handlers.NewScanHandler(repo)

	now := time.Now()
	job := &models.ScanJob{
		ID: "completed-scan-id", Type: models.ScanDiscovery,
		Target: "192.168.1.0/24", Status: models.ScanCompleted,
		StartedAt: &now, CompletedAt: &now,
	}
	repo.Create(context.Background(), job)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "completed-scan-id")
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/scans/completed-scan-id", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	h.Cancel(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}
