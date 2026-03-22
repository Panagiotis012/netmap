package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

// In-memory mock for testing
type mockDeviceRepo struct {
	devices map[string]*models.Device
}

func newMockDeviceRepo() *mockDeviceRepo {
	return &mockDeviceRepo{devices: make(map[string]*models.Device)}
}

func (m *mockDeviceRepo) List(ctx context.Context, params models.ListParams) (*models.ListResult[models.Device], error) {
	var items []models.Device
	for _, d := range m.devices {
		items = append(items, *d)
	}
	return &models.ListResult[models.Device]{Items: items, Total: len(items), Page: 1, TotalPages: 1}, nil
}

func (m *mockDeviceRepo) GetByID(ctx context.Context, id string) (*models.Device, error) {
	if d, ok := m.devices[id]; ok {
		return d, nil
	}
	return nil, store.ErrNotFound
}

func (m *mockDeviceRepo) GetByMAC(ctx context.Context, mac string) (*models.Device, error) {
	return nil, store.ErrNotFound
}
func (m *mockDeviceRepo) GetByHostname(ctx context.Context, hostname string) (*models.Device, error) {
	return nil, store.ErrNotFound
}
func (m *mockDeviceRepo) GetByIP(ctx context.Context, ip string) (*models.Device, error) {
	return nil, store.ErrNotFound
}
func (m *mockDeviceRepo) Create(ctx context.Context, d *models.Device) error {
	m.devices[d.ID] = d
	return nil
}
func (m *mockDeviceRepo) Update(ctx context.Context, d *models.Device) error {
	m.devices[d.ID] = d
	return nil
}
func (m *mockDeviceRepo) Delete(ctx context.Context, id string) error {
	delete(m.devices, id)
	return nil
}
func (m *mockDeviceRepo) UpdateStatus(ctx context.Context, id string, s models.DeviceStatus) error {
	return nil
}
func (m *mockDeviceRepo) UpdatePosition(ctx context.Context, id string, x, y float64) error {
	return nil
}
func (m *mockDeviceRepo) CountByStatus(ctx context.Context) (int, int, int, error) {
	return len(m.devices), 0, 0, nil
}

func TestListDevices(t *testing.T) {
	repo := newMockDeviceRepo()
	repo.devices["d1"] = &models.Device{
		ID: "d1", Hostname: "test", Status: models.StatusOnline,
		IPAddresses: []string{"192.168.1.10"}, Tags: []string{},
		FirstSeenAt: time.Now(), LastSeenAt: time.Now(),
	}

	h := NewDeviceHandler(repo)
	r := httptest.NewRequest("GET", "/api/v1/devices", nil)
	w := httptest.NewRecorder()
	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result models.ListResult[models.Device]
	json.NewDecoder(w.Body).Decode(&result)
	if result.Total != 1 {
		t.Errorf("expected 1, got %d", result.Total)
	}
}

func TestCreateDevice(t *testing.T) {
	repo := newMockDeviceRepo()
	h := NewDeviceHandler(repo)

	body := `{"hostname":"new-device","ip_addresses":["192.168.1.50"]}`
	r := httptest.NewRequest("POST", "/api/v1/devices", bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Create(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetDevice(t *testing.T) {
	repo := newMockDeviceRepo()
	repo.devices["d1"] = &models.Device{
		ID: "d1", Hostname: "test", Status: models.StatusOnline,
		Tags: []string{}, FirstSeenAt: time.Now(), LastSeenAt: time.Now(),
	}
	h := NewDeviceHandler(repo)

	r := httptest.NewRequest("GET", "/api/v1/devices/d1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "d1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	h.Get(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
