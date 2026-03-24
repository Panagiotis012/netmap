package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

var configDefaults = map[string]string{
	"scan_interval": "5m",
	"scan_workers":  "50",
	"port_ranges":   "22,80,443,8080,8443",
}

type ConfigRepo interface {
	Get(ctx context.Context, key string) string
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) map[string]string
}

type ConfigHandler struct {
	repo ConfigRepo
}

func NewConfigHandler(repo ConfigRepo) *ConfigHandler {
	return &ConfigHandler{repo: repo}
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.currentConfig(r.Context()))
}

func (h *ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {
	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	for k := range input {
		if _, ok := configDefaults[k]; !ok {
			writeError(w, http.StatusBadRequest, "unknown config key: "+k)
			return
		}
	}
	for k, v := range input {
		if err := h.repo.Set(r.Context(), k, v); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, h.currentConfig(r.Context()))
}

func (h *ConfigHandler) currentConfig(ctx context.Context) map[string]string {
	stored := h.repo.GetAll(ctx)
	result := make(map[string]string, len(configDefaults))
	for k, def := range configDefaults {
		if v, ok := stored[k]; ok {
			result[k] = v
		} else {
			result[k] = def
		}
	}
	return result
}
