package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

// MonitorRunner is the interface the handler uses to add/remove monitors from the runner.
type MonitorRunner interface {
	Add(m *models.Monitor)
	Remove(id string)
}

// MonitorRepoInterface is a local interface so the handler doesn't depend directly on store.MonitorRepo.
type MonitorRepoInterface interface {
	Create(ctx context.Context, m *models.Monitor) error
	List(ctx context.Context) ([]models.Monitor, error)
	GetByID(ctx context.Context, id string) (*models.Monitor, error)
	Update(ctx context.Context, m *models.Monitor) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status models.MonitorStatus, lastCheckedAt time.Time) error
	ListActive(ctx context.Context) ([]models.Monitor, error)
	CreateCheck(ctx context.Context, c *models.MonitorCheck) error
	ListChecks(ctx context.Context, monitorID string, limit int) ([]models.MonitorCheck, error)
	DeleteOldChecks(ctx context.Context, monitorID string, keepCount int) error
	UptimePercent(ctx context.Context, monitorID string, since time.Time) (float64, error)
}

// MonitorHandler handles HTTP requests for the monitors resource.
type MonitorHandler struct {
	repo   MonitorRepoInterface
	runner MonitorRunner
}

// NewMonitorHandler creates a new MonitorHandler.
func NewMonitorHandler(repo store.MonitorRepo, runner MonitorRunner) *MonitorHandler {
	return &MonitorHandler{repo: repo, runner: runner}
}

// List returns all monitors with uptime statistics.
func (h *MonitorHandler) List(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	for i := range monitors {
		day, _ := h.repo.UptimePercent(r.Context(), monitors[i].ID, now.Add(-24*time.Hour))
		week, _ := h.repo.UptimePercent(r.Context(), monitors[i].ID, now.Add(-7*24*time.Hour))
		monitors[i].UptimeDay = day
		monitors[i].UptimeWeek = week
	}

	writeJSON(w, http.StatusOK, monitors)
}

// Create creates a new monitor and registers it with the runner.
func (h *MonitorHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var m models.Monitor
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	m.ID = uuid.New().String()
	m.CreatedAt = time.Now()
	m.Status = models.MonitorStatusPending

	if m.Interval <= 0 {
		m.Interval = 60
	}
	if m.Timeout <= 0 {
		m.Timeout = 10
	}
	if m.Method == "" {
		m.Method = "GET"
	}
	if m.ExpectedStatus == 0 {
		m.ExpectedStatus = 200
	}

	if err := h.repo.Create(r.Context(), &m); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if m.Active {
		h.runner.Add(&m)
	}

	writeJSON(w, http.StatusCreated, m)
}

// Get returns a single monitor by ID.
func (h *MonitorHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	m, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "monitor not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	now := time.Now()
	m.UptimeDay, _ = h.repo.UptimePercent(r.Context(), m.ID, now.Add(-24*time.Hour))
	m.UptimeWeek, _ = h.repo.UptimePercent(r.Context(), m.ID, now.Add(-7*24*time.Hour))

	writeJSON(w, http.StatusOK, m)
}

// Update updates an existing monitor and restarts it in the runner.
func (h *MonitorHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	id := chi.URLParam(r, "id")
	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "monitor not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var input models.Monitor
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Apply updates.
	existing.Name = input.Name
	existing.Type = input.Type
	existing.URL = input.URL
	existing.Host = input.Host
	existing.Port = input.Port
	existing.Method = input.Method
	existing.ExpectedStatus = input.ExpectedStatus
	existing.Keyword = input.Keyword
	existing.Active = input.Active
	existing.NotifyWebhook = input.NotifyWebhook
	if input.Interval > 0 {
		existing.Interval = input.Interval
	}
	if input.Timeout > 0 {
		existing.Timeout = input.Timeout
	}

	if err := h.repo.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Restart in runner (Remove then Add if active).
	h.runner.Remove(existing.ID)
	if existing.Active {
		h.runner.Add(existing)
	}

	writeJSON(w, http.StatusOK, existing)
}

// Delete removes a monitor and stops it in the runner.
func (h *MonitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "monitor not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	h.runner.Remove(id)
	w.WriteHeader(http.StatusNoContent)
}

// ListChecks returns the most recent checks for a monitor.
func (h *MonitorHandler) ListChecks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	checks, err := h.repo.ListChecks(r.Context(), id, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if checks == nil {
		checks = []models.MonitorCheck{}
	}
	writeJSON(w, http.StatusOK, checks)
}

// StatusPage returns all active monitors with their current status (public endpoint).
func (h *MonitorHandler) StatusPage(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.repo.ListActive(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	for i := range monitors {
		day, _ := h.repo.UptimePercent(r.Context(), monitors[i].ID, now.Add(-24*time.Hour))
		week, _ := h.repo.UptimePercent(r.Context(), monitors[i].ID, now.Add(-7*24*time.Hour))
		monitors[i].UptimeDay = day
		monitors[i].UptimeWeek = week
	}

	if monitors == nil {
		monitors = []models.Monitor{}
	}
	writeJSON(w, http.StatusOK, monitors)
}
