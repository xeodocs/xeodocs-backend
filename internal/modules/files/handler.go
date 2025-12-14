package files

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
	r.Get("/projects/{projectId}/languages/{languageId}/files", h.ListFiles)
	r.Post("/projects/{projectId}/languages/{languageId}/files", h.CreateFile)
	r.Get("/projects/{projectId}/languages/{languageId}/files/{fileId}", h.GetFile)
	r.Patch("/projects/{projectId}/languages/{languageId}/files/{fileId}", h.UpdateFile)
	r.Delete("/projects/{projectId}/languages/{languageId}/files/{fileId}", h.DeleteFile)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	languageID := chi.URLParam(r, "languageId")
	files, err := h.repo.GetAll(languageID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if files == nil {
		files = []File{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": files})
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "fileId")
	file, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "File not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": file})
}

func (h *Handler) CreateFile(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	languageID := chi.URLParam(r, "languageId")

	var f File
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if f.Path == "" {
		response.Error(w, http.StatusBadRequest, "Missing required fields", nil)
		return
	}

	f.ProjectID = projectID
	f.LanguageID = languageID
	if f.Status == "" {
		f.Status = "pending"
	}

	if err := h.repo.Create(&f); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusCreated, map[string]interface{}{"data": f})
}

func (h *Handler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "fileId")

	existing, err := h.repo.GetByID(id)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "File not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var f File
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if f.Path != "" {
		existing.Path = f.Path
	}
	if f.Status != "" {
		existing.Status = f.Status
	}
	// We might allow updating checksums if manual, but usually automated.
	if f.ChecksumOriginal != "" {
		existing.ChecksumOriginal = f.ChecksumOriginal
	}
	if f.ChecksumTranslated != "" {
		existing.ChecksumTranslated = f.ChecksumTranslated
	}

	if err := h.repo.Update(id, existing); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": existing})
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "fileId")
	if err := h.repo.Delete(id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
