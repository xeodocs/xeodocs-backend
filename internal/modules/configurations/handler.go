package configurations

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/configurations", h.ListConfigurations)
	r.Post("/configurations", h.CreateConfiguration)
	r.Get("/configurations/{key}", h.GetConfiguration)
	r.Put("/configurations/{key}", h.UpsertConfiguration)
	r.Patch("/configurations/{key}", h.UpdateConfiguration)
	r.Delete("/configurations/{key}", h.DeleteConfiguration)
}

func (h *Handler) ListConfigurations(w http.ResponseWriter, r *http.Request) {
	configs, err := h.repo.GetAll()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if configs == nil {
		configs = []Configuration{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": configs})
}

func (h *Handler) CreateConfiguration(w http.ResponseWriter, r *http.Request) {
	var c Configuration
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if c.Key == "" {
		response.Error(w, http.StatusBadRequest, "Key is required", nil)
		return
	}

	if err := h.repo.Upsert(&c); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": c})
}

func (h *Handler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	config, err := h.repo.GetByKey(key)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Configuration not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": config})
}

func (h *Handler) UpsertConfiguration(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var c Configuration
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	c.Key = key

	if err := h.repo.Upsert(&c); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": c})
}

func (h *Handler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	// Check existence
	_, err := h.repo.GetByKey(key)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Configuration not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var c Configuration
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	c.Key = key

	if err := h.repo.Upsert(&c); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": c})
}

func (h *Handler) DeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if err := h.repo.Delete(key); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
