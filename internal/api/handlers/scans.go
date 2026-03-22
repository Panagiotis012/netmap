package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type ScanHandler struct {
	repo store.ScanRepo
	// ScanTrigger will be set by the server to trigger actual scans
	ScanTrigger func(scanType models.ScanType, target string)
}

func NewScanHandler(repo store.ScanRepo) *ScanHandler {
	return &ScanHandler{repo: repo}
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
		writeError(w, http.StatusNotFound, "scan not found")
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
	if h.ScanTrigger != nil {
		go h.ScanTrigger(input.Type, input.Target)
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "scan triggered"})
}
