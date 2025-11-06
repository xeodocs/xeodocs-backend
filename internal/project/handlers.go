package project

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/xeodocs/xeodocs-backend/internal/shared/auth"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func CreateProjectHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		var req CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if req.RepoURL == "" {
			http.Error(w, "repo_url is required", http.StatusBadRequest)
			return
		}

		project, err := CreateProject(claims.UserID, req)
		if err != nil {
			log.Println("Error creating project:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(project)
	}
}

func ListProjectsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		projects, err := GetProjects(claims.UserID)
		if err != nil {
			log.Println("Error getting projects:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

func GetProjectHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		idStr := r.URL.Path[len("/projects/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		project, err := GetProjectByID(id, claims.UserID)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error getting project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}
}

func UpdateProjectHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		idStr := r.URL.Path[len("/projects/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		var req UpdateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		project, err := UpdateProject(id, claims.UserID, req)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error updating project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(project)
	}
}

func DeleteProjectHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		idStr := r.URL.Path[len("/projects/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		err = DeleteProject(id, claims.UserID)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error deleting project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
