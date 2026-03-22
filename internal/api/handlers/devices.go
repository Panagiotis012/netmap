package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type DeviceHandler struct {
	repo store.DeviceRepo
}

func NewDeviceHandler(repo store.DeviceRepo) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	params := models.ListParams{
		Page:   queryInt(r, "page", 1),
		Limit:  queryInt(r, "limit", 50),
		Sort:   r.URL.Query().Get("sort"),
		Order:  r.URL.Query().Get("order"),
		Search: r.URL.Query().Get("search"),
		Status: models.DeviceStatus(r.URL.Query().Get("status")),
	}

	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	device, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var input struct {
		Hostname    string   `json:"hostname"`
		IPAddresses []string `json:"ip_addresses"`
		Tags        []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	device := &models.Device{
		ID:              uuid.New().String(),
		Hostname:        input.Hostname,
		IPAddresses:     input.IPAddresses,
		MACAddresses:    []string{},
		Tags:            input.Tags,
		Status:          models.StatusUnknown,
		DiscoveryMethod: models.DiscoveryManual,
		FirstSeenAt:     time.Now(),
		LastSeenAt:      time.Now(),
	}
	if device.Tags == nil {
		device.Tags = []string{}
	}

	if err := h.repo.Create(r.Context(), device); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, device)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	id := chi.URLParam(r, "id")
	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if v, ok := input["hostname"].(string); ok {
		existing.Hostname = v
	}
	if v, ok := input["tags"].([]interface{}); ok {
		tags := make([]string, len(v))
		for i, t := range v {
			tags[i], _ = t.(string)
		}
		existing.Tags = tags
	}
	if v, ok := input["map_x"].(float64); ok {
		existing.MapX = &v
	}
	if v, ok := input["map_y"].(float64); ok {
		existing.MapY = &v
	}

	existing.LastSeenAt = time.Now()
	if err := h.repo.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return i
}
