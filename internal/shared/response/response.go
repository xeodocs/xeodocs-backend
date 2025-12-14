package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func Error(w http.ResponseWriter, status int, message string, details []ErrorDetail) {
	resp := ErrorResponse{
		Code:    status,
		Message: message,
		Details: details,
	}
	JSON(w, status, resp)
}
