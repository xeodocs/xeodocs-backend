package languages

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xeodocs/xeodocs-backend/internal/modules/projects"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	gh "github.com/xeodocs/xeodocs-backend/internal/shared/github"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type Handler struct {
	repo         *Repository
	projectsRepo *projects.Repository
	ghService    *gh.Service
	config       *config.Config
}

func NewHandler(repo *Repository, projectsRepo *projects.Repository, ghService *gh.Service, cfg *config.Config) *Handler {
	return &Handler{
		repo:         repo,
		projectsRepo: projectsRepo,
		ghService:    ghService,
		config:       cfg,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/languages", h.ListAllLanguages)
	r.Post("/languages", h.CreateLanguage)
	r.Get("/languages/{languageId}", h.GetLanguage)
	r.Patch("/languages/{languageId}", h.UpdateLanguage)
	r.Delete("/languages/{languageId}", h.DeleteLanguage)

	r.Get("/projects/{projectId}/languages", h.ListLanguages)
	r.Post("/projects/{projectId}/languages", h.CreateLanguage)
	r.Get("/projects/{projectId}/languages/{languageId}", h.GetLanguage)
	r.Patch("/projects/{projectId}/languages/{languageId}", h.UpdateLanguage)
	r.Delete("/projects/{projectId}/languages/{languageId}", h.DeleteLanguage)
}

func (h *Handler) ListAllLanguages(w http.ResponseWriter, r *http.Request) {
	languages, err := h.repo.GetAll("")
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if languages == nil {
		languages = []Language{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": languages})
}

func (h *Handler) ListLanguages(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	languages, err := h.repo.GetAll(projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if languages == nil {
		languages = []Language{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": languages})
}

func (h *Handler) GetLanguage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "languageId")
	language, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Language not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": language})
}

func (h *Handler) CreateLanguage(w http.ResponseWriter, r *http.Request) {
	var l Language
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// If projectId is in the URL, it takes precedence
	if urlProjectID := chi.URLParam(r, "projectId"); urlProjectID != "" {
		l.ProjectID = urlProjectID
	}

	if l.ProjectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing required field: projectId", nil)
		return
	}

	if l.Code == "" || l.Name == "" {
		response.Error(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	l.IsActive = true

	// Ensure branches exist in the fork
	// 1. Get Project to find slug and source branch
	project, err := h.projectsRepo.GetByID(l.ProjectID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Error(w, http.StatusNotFound, "Project not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project: "+err.Error(), nil)
		return
	}

	// 2. Determine fork name
	forkName := h.getForkName(project.Slug)

	// 3. Ensure [code] branch exists (based on source branch)
	if err := h.ghService.EnsureBranchExists(forkName, l.Code, project.SourceBranch); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create language branch: "+err.Error(), nil)
		return
	}

	// 4. Ensure local-[code] branch exists (based on source branch)
	if err := h.ghService.EnsureBranchExists(forkName, "local-"+l.Code, project.SourceBranch); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create local language branch: "+err.Error(), nil)
		return
	}

	if err := h.repo.Create(&l); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{"data": l})
}

func (h *Handler) UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "languageId")

	existing, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Language not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var l Language
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if l.Code != "" {
		existing.Code = l.Code
	}
	if l.Name != "" {
		existing.Name = l.Name
	}
	if l.Domain != "" {
		existing.Domain = l.Domain
	}
	existing.IsActive = l.IsActive

	// Ensure branches exist (in case they didn't, or code changed)
	// Even if code changed, we create new branches for the new code.
	// We don't delete old ones automatically.

	// 1. Get Project
	project, err := h.projectsRepo.GetByID(existing.ProjectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch project: "+err.Error(), nil)
		return
	}

	// 2. Fork Name
	forkName := h.getForkName(project.Slug)

	// 3. Ensure branches
	if err := h.ghService.EnsureBranchExists(forkName, existing.Code, project.SourceBranch); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to verify/create language branch: "+err.Error(), nil)
		return
	}
	if err := h.ghService.EnsureBranchExists(forkName, "local-"+existing.Code, project.SourceBranch); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to verify/create local language branch: "+err.Error(), nil)
		return
	}

	if err := h.repo.Update(id, existing); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": existing})
}

func (h *Handler) getForkName(slug string) string {
	switch h.config.Environment {
	case "development":
		return "development-" + slug
	case "staging":
		return "staging-" + slug
	default: // production
		return slug
	}
}

func (h *Handler) DeleteLanguage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "languageId")
	if err := h.repo.Delete(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
