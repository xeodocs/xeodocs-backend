package logging

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
)

// Log represents a log entry
type Log struct {
	ID        int       `json:"logId" db:"id"`
	Type      string    `json:"type" db:"type"`
	Message   string    `json:"message" db:"message"`
	UserID    *int      `json:"userId,omitempty" db:"user_id"`
	ProjectID *int      `json:"projectId,omitempty" db:"project_id"`
	CreatedAt time.Time `json:"timestamp" db:"created_at"`
	Level     string    `json:"level" db:"level"`
}

// CreateLogRequest represents the request to create a new log
type CreateLogRequest struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	UserID    *int   `json:"userId,omitempty"`
	ProjectID *int   `json:"projectId,omitempty"`
	Level     string `json:"level"`
}

// QueryLogsResponse represents the response for querying logs
type QueryLogsResponse struct {
	Logs  []Log `json:"logs"`
	Total int   `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// CreateLog inserts a new log entry into the database
func CreateLog(req CreateLogRequest) (*Log, error) {
	log := &Log{
		Type:      req.Type,
		Message:   req.Message,
		UserID:    req.UserID,
		ProjectID: req.ProjectID,
		Level:     req.Level,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO logs (type, message, user_id, project_id, level, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := db.DB.QueryRow(query, log.Type, log.Message, log.UserID, log.ProjectID, log.Level, log.CreatedAt).Scan(&log.ID)
	if err != nil {
		return nil, err
	}

	return log, nil
}

// QueryLogs retrieves logs with optional filters and pagination
func QueryLogs(logType, userIDStr, projectIDStr, level string, page, limit int) ([]Log, int, error) {
	var logs []Log
	var total int

	// Base query
	query := `SELECT id, type, message, user_id, project_id, created_at, level FROM logs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM logs WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	if logType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, logType)
		argIndex++
	}

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return nil, 0, errors.New("invalid userId")
		}
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if projectIDStr != "" {
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			return nil, 0, errors.New("invalid projectId")
		}
		query += fmt.Sprintf(" AND project_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND project_id = $%d", argIndex)
		args = append(args, projectID)
		argIndex++
	}

	if level != "" {
		query += fmt.Sprintf(" AND level = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND level = $%d", argIndex)
		args = append(args, level)
		argIndex++
	}

	// Get total count
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var log Log
		err := rows.Scan(&log.ID, &log.Type, &log.Message, &log.UserID, &log.ProjectID, &log.CreatedAt, &log.Level)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}
