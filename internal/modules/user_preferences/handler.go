package user_preferences

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
	// Assuming routes are usually under /users/{userId}/preferences or similar
	// But spec said "UserPreferences" tag. Let's assume /users/{userId}/preferences based on standard practice and the memory mention.
	r.Get("/users/{userId}/preferences", h.ListPreferences)
	r.Get("/users/{userId}/preferences/{key}", h.GetPreference)
	r.Put("/users/{userId}/preferences/{key}", h.UpsertPreference)
	r.Delete("/users/{userId}/preferences/{key}", h.DeletePreference)
}

func (h *Handler) ListPreferences(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	prefs, err := h.repo.GetAll(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if prefs == nil {
		prefs = []UserPreference{}
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": prefs})
}

func (h *Handler) GetPreference(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	key := chi.URLParam(r, "key")

	pref, err := h.repo.GetByKey(userID, key)
	if err == sql.ErrNoRows {
		response.Error(w, http.StatusNotFound, "Preference not found", nil)
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": pref})
}

func (h *Handler) UpsertPreference(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	key := chi.URLParam(r, "key")

	var p UserPreference
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	p.UserID = userID
	p.Key = key

	if err := h.repo.Upsert(&p); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{"data": p})
}

func (h *Handler) DeletePreference(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	key := chi.URLParam(r, "key")
	if err := h.repo.Delete(userID, key); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
