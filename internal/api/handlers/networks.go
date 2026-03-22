package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

type NetworkHandler struct {
	repo store.NetworkRepo
}

func NewNetworkHandler(repo store.NetworkRepo) *NetworkHandler {
	return &NetworkHandler{repo: repo}
}

func (h *NetworkHandler) List(w http.ResponseWriter, r *http.Request) {
	nets, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nets)
}

func (h *NetworkHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	net, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "network not found")
		return
	}
	writeJSON(w, http.StatusOK, net)
}

func (h *NetworkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		Subnet  string `json:"subnet"`
		Gateway string `json:"gateway"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	now := time.Now()
	net := &models.Network{
		ID: uuid.New().String(), Name: input.Name,
		Subnet: input.Subnet, Gateway: input.Gateway,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := h.repo.Create(r.Context(), net); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, net)
}

func (h *NetworkHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	existing, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "network not found")
		return
	}
	var input struct {
		Name    *string `json:"name"`
		Subnet  *string `json:"subnet"`
		Gateway *string `json:"gateway"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Subnet != nil {
		existing.Subnet = *input.Subnet
	}
	if input.Gateway != nil {
		existing.Gateway = *input.Gateway
	}
	existing.UpdatedAt = time.Now()
	if err := h.repo.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (h *NetworkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
