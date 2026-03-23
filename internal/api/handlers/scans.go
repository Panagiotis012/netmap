package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type ScanHandler struct {
	repo        store.ScanRepo
	mu          sync.Mutex
	cancels     map[string]context.CancelFunc
	ScanTrigger func(ctx context.Context, scanID string, scanType models.ScanType, target string)
}

func NewScanHandler(repo store.ScanRepo) *ScanHandler {
	return &ScanHandler{
		repo:    repo,
		cancels: make(map[string]context.CancelFunc),
	}
}

func (h *ScanHandler) List(w http.ResponseWriter, r *http.Request) {
	params := models.ListParams{
		Page:  queryInt(r, "page", 1),
		Limit: queryInt(r, "limit", 50),
	}
	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ScanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	scan, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "scan not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, scan)
}

func (h *ScanHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type   models.ScanType `json:"type"`
		Target string          `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.Type == "" {
		input.Type = models.ScanDiscovery
	}
	if input.Target == "" {
		writeError(w, http.StatusBadRequest, "target is required")
		return
	}

	h.mu.Lock()
	if len(h.cancels) > 0 {
		h.mu.Unlock()
		writeError(w, http.StatusConflict, "scan already in progress")
		return
	}

	scanID := uuid.New().String()
	now := time.Now()
	job := &models.ScanJob{
		ID:        scanID,
		Type:      input.Type,
		Target:    input.Target,
		Status:    models.ScanRunning,
		StartedAt: &now,
	}
	if err := h.repo.Create(r.Context(), job); err != nil {
		h.mu.Unlock()
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.cancels[scanID] = cancel
	h.mu.Unlock()

	if h.ScanTrigger != nil {
		go func() {
			defer func() {
				h.mu.Lock()
				delete(h.cancels, scanID)
				h.mu.Unlock()
			}()
			h.ScanTrigger(ctx, scanID, input.Type, input.Target)
		}()
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"id": scanID, "status": "running"})
}

func (h *ScanHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := h.repo.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "scan not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.mu.Lock()
	cancel, running := h.cancels[id]
	if running {
		delete(h.cancels, id)
	}
	h.mu.Unlock()

	if !running {
		writeError(w, http.StatusConflict, "scan not running")
		return
	}

	cancel()
	w.WriteHeader(http.StatusNoContent)
}
