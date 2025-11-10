package project

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/xeodocs/xeodocs-backend/internal/shared/auth"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) *int {
	claims, ok := ctx.Value("claims").(*auth.Claims)
	if !ok {
		return nil
	}
	return &claims.UserID
}

func CreateProjectHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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
		if req.DocURL == "" {
			http.Error(w, "doc_url is required", http.StatusBadRequest)
			return
		}
		if req.RepoURL == "" {
			http.Error(w, "repo_url is required", http.StatusBadRequest)
			return
		}
		if req.BuildCommand == "" {
			http.Error(w, "build_command is required", http.StatusBadRequest)
			return
		}

		project, err := CreateProject(req)
		if err != nil {
			log.Println("Error creating project:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the project creation
		userID := getUserIDFromContext(r.Context())
		message := "Project created: " + project.Name
		logging.LogActivity(cfg.LoggingServiceURL, "project_created", message, userID, &project.ID, "info")

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

		projects, err := GetProjects()
		if err != nil {
			log.Println("Error getting projects:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the project listing
		userID := getUserIDFromContext(r.Context())
		message := "Projects listed"
		logging.LogActivity(cfg.LoggingServiceURL, "projects_listed", message, userID, nil, "info")

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

		idStr := r.URL.Path[len("/projects/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		project, err := GetProjectByID(id)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error getting project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Log the project retrieval
		userID := getUserIDFromContext(r.Context())
		message := "Project retrieved: " + project.Name
		logging.LogActivity(cfg.LoggingServiceURL, "project_retrieved", message, userID, &project.ID, "info")

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

		project, err := UpdateProject(id, req)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error updating project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Log the project update
		userID := getUserIDFromContext(r.Context())
		message := "Project updated: " + project.Name
		logging.LogActivity(cfg.LoggingServiceURL, "project_updated", message, userID, &project.ID, "info")

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

		idStr := r.URL.Path[len("/projects/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		err = DeleteProject(id)
		if err != nil {
			if err.Error() == "project not found" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				log.Println("Error deleting project:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Log the project deletion
		userID := getUserIDFromContext(r.Context())
		message := "Project deleted with ID: " + strconv.Itoa(id)
		logging.LogActivity(cfg.LoggingServiceURL, "project_deleted", message, userID, &id, "info")

		w.WriteHeader(http.StatusNoContent)
	}
}
