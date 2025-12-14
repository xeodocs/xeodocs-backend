package system

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/response"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.HealthCheck)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": "1.0.0",
	})
}
