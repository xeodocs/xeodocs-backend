package paths

import (
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
	// Ignored Paths
	r.Get("/projects/{projectId}/ignored-paths", h.ListIgnoredPaths)
	r.Post("/projects/{projectId}/ignored-paths", h.CreateIgnoredPath)
	r.Delete("/projects/{projectId}/ignored-paths/{id}", h.DeleteIgnoredPath)

	// Special Paths
	r.Get("/projects/{projectId}/special-paths", h.ListSpecialPaths)
	r.Post("/projects/{projectId}/special-paths", h.CreateSpecialPath)
	r.Delete("/projects/{projectId}/special-paths/{id}", h.DeleteSpecialPath)
}

// Ignored Paths

func (h *Handler) ListIgnoredPaths(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	paths, err := h.repo.GetIgnoredPaths(projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if paths == nil {
		paths = []IgnoredPath{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": paths})
}

func (h *Handler) CreateIgnoredPath(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	var p IgnoredPath
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	p.ProjectID = projectID

	if err := h.repo.CreateIgnoredPath(&p); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{"data": p})
}

func (h *Handler) DeleteIgnoredPath(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.DeleteIgnoredPath(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Special Paths

func (h *Handler) ListSpecialPaths(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	paths, err := h.repo.GetSpecialPaths(projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if paths == nil {
		paths = []SpecialPath{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": paths})
}

func (h *Handler) CreateSpecialPath(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	var p SpecialPath
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	p.ProjectID = projectID

	if err := h.repo.CreateSpecialPath(&p); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{"data": p})
}

func (h *Handler) DeleteSpecialPath(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.DeleteSpecialPath(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
