package build

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

// BuildHandler handles build requests for projects
func BuildHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ProjectID int `json:"projectId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log build start
		message := "Build service starting build for project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "build_start", message, nil, &req.ProjectID, "info")

		// Execute build
		err := ExecuteBuild(req.ProjectID, cfg)
		if err != nil {
			message := "Build service failed to build project " + strconv.Itoa(req.ProjectID) + ": " + err.Error()
			logging.LogActivity(cfg.LoggingServiceURL, "build_error", message, nil, &req.ProjectID, "error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log build success
		message = "Build service successfully built project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "build_success", message, nil, &req.ProjectID, "info")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// ExportHandler handles export requests for projects
func ExportHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ProjectID int `json:"projectId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log export start
		message := "Build service starting export for project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "export_start", message, nil, &req.ProjectID, "info")

		// Execute export
		err := ExecuteExport(req.ProjectID, cfg)
		if err != nil {
			message := "Build service failed to export project " + strconv.Itoa(req.ProjectID) + ": " + err.Error()
			logging.LogActivity(cfg.LoggingServiceURL, "export_error", message, nil, &req.ProjectID, "error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log export success
		message = "Build service successfully exported project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "export_success", message, nil, &req.ProjectID, "info")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// PreviewHandler handles preview requests for projects
func PreviewHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ProjectID int `json:"projectId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log preview start
		message := "Build service starting preview for project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "preview_start", message, nil, &req.ProjectID, "info")

		// Execute preview
		err := ExecutePreview(req.ProjectID, cfg)
		if err != nil {
			message := "Build service failed to preview project " + strconv.Itoa(req.ProjectID) + ": " + err.Error()
			logging.LogActivity(cfg.LoggingServiceURL, "preview_error", message, nil, &req.ProjectID, "error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log preview success
		message = "Build service successfully started preview for project " + strconv.Itoa(req.ProjectID)
		logging.LogActivity(cfg.LoggingServiceURL, "preview_success", message, nil, &req.ProjectID, "info")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
