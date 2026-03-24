package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/netmap/netmap/internal/core/models"
)

type AlertRepo interface {
	Create(ctx context.Context, alert *models.Alert) error
	List(ctx context.Context, limit int) ([]models.Alert, error)
	MarkAllRead(ctx context.Context) error
	DeleteAll(ctx context.Context) error
	UnreadCount(ctx context.Context) (int, error)
	Trim(ctx context.Context, keep int) error
}

type AlertHandler struct {
	repo AlertRepo
}

func NewAlertHandler(repo AlertRepo) *AlertHandler {
	return &AlertHandler{repo: repo}
}

func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	var a models.Alert
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if a.ID == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	if a.Timestamp.IsZero() {
		a.Timestamp = time.Now()
	}
	if err := h.repo.Create(r.Context(), &a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Keep only the most recent 500 alerts.
	_ = h.repo.Trim(r.Context(), 500)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.repo.List(r.Context(), 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	unread, _ := h.repo.UnreadCount(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"unread": unread,
	})
}

func (h *AlertHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.MarkAllRead(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertHandler) DeleteAll(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteAll(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
