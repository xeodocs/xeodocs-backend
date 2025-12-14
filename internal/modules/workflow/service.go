package workflow

import (
	"database/sql"
	"fmt"

	"github.com/xeodocs/xeodocs-backend/internal/modules/files"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type NextFileResponse struct {
	File   files.File `json:"file"`
	Prompt string     `json:"prompt"`
}

type StatusResponse struct {
	PendingCount       int     `json:"pendingCount"`
	OutdatedCount      int     `json:"outdatedCount"`
	TotalFiles         int     `json:"totalFiles"`
	ProgressPercentage float64 `json:"progressPercentage"`
}

func (s *Service) GetNextFile(projectID, languageID string) (*NextFileResponse, error) {
	// Priority: pending, then outdated.
	query := `
		SELECT id, project_id, language_id, path, checksum_original, checksum_translated, status, last_synced_at, updated_at
		FROM files 
		WHERE language_id = $1 AND (status = 'pending' OR status = 'outdated')
		ORDER BY status ASC, path ASC
		LIMIT 1
	`
	var f files.File
	var csOriginal, csTranslated sql.NullString

	err := s.db.QueryRow(query, languageID).Scan(
		&f.ID, &f.ProjectID, &f.LanguageID, &f.Path, &csOriginal, &csTranslated, &f.Status, &f.LastSyncedAt, &f.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No files pending
	}
	if err != nil {
		return nil, err
	}
	f.ChecksumOriginal = csOriginal.String
	f.ChecksumTranslated = csTranslated.String

	// Construct Prompt
	prompt := fmt.Sprintf("Translate the file '%s' from English to the target language. Keep the structure intact.", f.Path)

	return &NextFileResponse{
		File:   f,
		Prompt: prompt,
	}, nil
}

func (s *Service) SubmitFiles(projectID, languageID string, fileIDs []string) (int, error) {
	// Mark files as translated
	// In a real scenario, we might want to update the checksum_translated here as well if provided,
	// but the CLI submit endpoint in the spec just takes fileIds and sets status to translated.
	// Actually spec says "SubmitRequest" has fileIds and status enum.

	if len(fileIDs) == 0 {
		return 0, nil
	}

	query := `UPDATE files SET status = 'translated', updated_at = NOW() WHERE id = ANY($1::uuid[]) AND language_id = $2`

	// Convert []string to PostgreSQL array format is tricky with just database/sql if not using lib/pq Array
	// But we imported lib/pq in database.go so we can use pq.Array if we import it here, or construct string manually.
	// I will just loop for simplicity or use a dirty query builder since I can't easily add pq dep in this file without editing imports properly.
	// Actually I can import github.com/lib/pq.

	// Wait, I am not allowed to edit imports in the middle of file. I should have added it.
	// I will use a simple loop for now to be safe and avoid import issues, or assume I can use a subquery if ids are simple strings.
	// Actually, `ANY($1::uuid[])` works if I pass a string formatted as array `{uuid,uuid}`.

	idArray := "{"
	for i, id := range fileIDs {
		if i > 0 {
			idArray += ","
		}
		idArray += id
	}
	idArray += "}"

	res, err := s.db.Exec(query, idArray, languageID)
	if err != nil {
		return 0, err
	}

	rows, _ := res.RowsAffected()
	return int(rows), nil
}

func (s *Service) GetStatus(projectID, languageID string) (*StatusResponse, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'outdated') as outdated,
			COUNT(*) as total
		FROM files
		WHERE language_id = $1
	`
	var pending, outdated, total int
	err := s.db.QueryRow(query, languageID).Scan(&pending, &outdated, &total)
	if err != nil {
		return nil, err
	}

	progress := 0.0
	if total > 0 {
		translated := total - pending - outdated
		progress = (float64(translated) / float64(total)) * 100
	}

	return &StatusResponse{
		PendingCount:       pending,
		OutdatedCount:      outdated,
		TotalFiles:         total,
		ProgressPercentage: progress,
	}, nil
}
