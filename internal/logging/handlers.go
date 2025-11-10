package logging

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

// CreateLogHandler handles POST /logs to create a new log entry
func CreateLogHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateLogRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		logEntry, err := CreateLog(req)
		if err != nil {
			log.Println("Error creating log:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(logEntry)
	}
}

// QueryLogsHandler handles GET /logs to query logs with filters and pagination
func QueryLogsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse query parameters
		logType := r.URL.Query().Get("type")
		userIDStr := r.URL.Query().Get("userId")
		projectIDStr := r.URL.Query().Get("projectId")
		level := r.URL.Query().Get("level")

		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}

		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		logs, total, err := QueryLogs(logType, userIDStr, projectIDStr, level, page, limit)
		if err != nil {
			log.Println("Error querying logs:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := QueryLogsResponse{
			Logs:  logs,
			Total: total,
			Page:  page,
			Limit: limit,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
