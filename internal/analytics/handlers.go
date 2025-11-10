package analytics

import (
	"encoding/json"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

type TrackEvent struct {
	Event     string                 `json:"event"`
	UserID    *int                   `json:"userId,omitempty"`
	ProjectID *int                   `json:"projectId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func TrackHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event TrackEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Log the event
		message := "Analytics event: " + event.Event
		logging.LogActivity(cfg.LoggingServiceURL, "analytics_event", message, event.UserID, event.ProjectID, "info")

		// TODO: Send to Kafka and InfluxDB

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
