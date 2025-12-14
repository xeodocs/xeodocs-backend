package languages

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

	if err := h.repo.Update(id, existing); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": existing})
}

func (h *Handler) DeleteLanguage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "languageId")
	if err := h.repo.Delete(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
