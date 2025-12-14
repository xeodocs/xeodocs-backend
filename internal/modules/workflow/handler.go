package workflow

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/projects/{projectId}/languages/{languageId}/next-file", h.GetNextFile)
	r.Post("/projects/{projectId}/languages/{languageId}/submissions", h.SubmitFiles)
	r.Get("/projects/{projectId}/languages/{languageId}/status", h.GetStatus)
}

func (h *Handler) GetNextFile(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	languageID := chi.URLParam(r, "languageId")

	resp, err := h.service.GetNextFile(projectID, languageID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if resp == nil {
		response.Error(w, http.StatusNotFound, "No files pending translation", nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": resp})
}

func (h *Handler) SubmitFiles(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	languageID := chi.URLParam(r, "languageId")

	var req struct {
		FileIDs []string `json:"fileIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	count, err := h.service.SubmitFiles(projectID, languageID, req.FileIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"updatedCount": count,
		},
	})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")
	languageID := chi.URLParam(r, "languageId")

	status, err := h.service.GetStatus(projectID, languageID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"data": status})
}
