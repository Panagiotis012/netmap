package handlers

import (
	"net/http"
	"runtime"

	"github.com/netmap/netmap/internal/store"
)

type SystemHandler struct {
	devices store.DeviceRepo
	version string
}

func NewSystemHandler(devices store.DeviceRepo, version string) *SystemHandler {
	return &SystemHandler{devices: devices, version: version}
}

func (h *SystemHandler) Status(w http.ResponseWriter, r *http.Request) {
	online, offline, unknown, _ := h.devices.CountByStatus(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"version":         h.version,
		"go_version":      runtime.Version(),
		"devices_online":  online,
		"devices_offline": offline,
		"devices_unknown": unknown,
		"devices_total":   online + offline + unknown,
	})
}
