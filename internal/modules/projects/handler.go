package projects

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	gh "github.com/xeodocs/xeodocs-backend/internal/shared/github"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type Handler struct {
	repo      *Repository
	ghService *gh.Service
	config    *config.Config
}

func NewHandler(repo *Repository, ghService *gh.Service, cfg *config.Config) *Handler {
	return &Handler{
		repo:      repo,
		ghService: ghService,
		config:    cfg,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/projects", h.ListProjects)
	r.Post("/projects", h.CreateProject)
	r.Get("/projects/{projectId}", h.GetProject)
	r.Patch("/projects/{projectId}", h.UpdateProject)
	r.Delete("/projects/{projectId}", h.DeleteProject)
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.repo.GetAll()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if projects == nil {
		projects = []Project{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": projects})
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")
	project, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Project not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": project})
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Basic validation
	if p.Name == "" || p.Slug == "" || p.SourceRepoURL == "" {
		response.Error(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}
	if p.SourceBranch == "" {
		p.SourceBranch = "main"
	}
	p.IsActive = true

	// Ensure Fork Exists
	forkName := h.getForkName(p.Slug)
	forkURL, err := h.ghService.EnsureForkExists(p.SourceRepoURL, forkName)
	if err != nil {
		// We fail the creation if we cannot fork, as per requirements
		response.Error(w, http.StatusInternalServerError, "Failed to create/verify GitHub fork: "+err.Error(), nil)
		return
	}

	if err := h.repo.Create(&p); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Add fork info to response
	respData := map[string]interface{}{
		"data": p,
		"meta": map[string]string{
			"forkUrl": forkURL,
		},
	}

	response.JSON(w, http.StatusCreated, respData)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")

	// First check if exists
	existing, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Project not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var p Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Merge updates
	if p.Name != "" {
		existing.Name = p.Name
	}
	if p.Slug != "" {
		existing.Slug = p.Slug
	}
	if p.SourceWebsiteURL != "" {
		existing.SourceWebsiteURL = p.SourceWebsiteURL
	}
	if p.SourceRepoURL != "" {
		existing.SourceRepoURL = p.SourceRepoURL
	}
	if p.SourceBranch != "" {
		existing.SourceBranch = p.SourceBranch
	}
	if p.Description != "" {
		existing.Description = p.Description
	}
	existing.IsActive = p.IsActive

	existing.Name = nonEmpty(p.Name, existing.Name)
	existing.Slug = nonEmpty(p.Slug, existing.Slug)
	existing.SourceWebsiteURL = nonEmpty(p.SourceWebsiteURL, existing.SourceWebsiteURL)
	existing.SourceRepoURL = nonEmpty(p.SourceRepoURL, existing.SourceRepoURL)
	existing.SourceBranch = nonEmpty(p.SourceBranch, existing.SourceBranch)
	existing.Description = nonEmpty(p.Description, existing.Description)
	existing.IsActive = p.IsActive

	// Ensure Fork Exists (even on edit)
	forkName := h.getForkName(existing.Slug)
	forkURL, err := h.ghService.EnsureForkExists(existing.SourceRepoURL, forkName)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to verify/create GitHub fork: "+err.Error(), nil)
		return
	}

	if err := h.repo.Update(id, existing); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	respData := map[string]interface{}{
		"data": existing,
		"meta": map[string]string{
			"forkUrl": forkURL,
		},
	}
	response.JSON(w, http.StatusOK, respData)
}

func (h *Handler) getForkName(slug string) string {
	switch h.config.Environment {
	case "development":
		return "development-" + slug
	case "staging":
		return "staging-" + slug
	default: // production or others
		return slug
	}
}

func nonEmpty(newVal, oldVal string) string {
	if newVal != "" {
		return newVal
	}
	return oldVal
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "projectId")
	if err := h.repo.Delete(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
