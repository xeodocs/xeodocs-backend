package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// LogActivity sends a log entry to the logging service
func LogActivity(loggingServiceURL, logType, message string, userID, projectID *int, level string) error {
	req := map[string]interface{}{
		"type":    logType,
		"message": message,
		"level":   level,
	}
	if userID != nil {
		req["userId"] = *userID
	}
	if projectID != nil {
		req["projectId"] = *projectID
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal log request: %w", err)
	}

	resp, err := http.Post(loggingServiceURL+"/logs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("logging service returned status %d", resp.StatusCode)
	}

	return nil
}
