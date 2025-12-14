package files

import (
	"database/sql"
	"time"
)

type File struct {
	ID                 string    `json:"id"`
	ProjectID          string    `json:"projectId"`
	LanguageID         string    `json:"languageId"`
	Path               string    `json:"path"`
	ChecksumOriginal   string    `json:"checksumOriginal"`
	ChecksumTranslated string    `json:"checksumTranslated"`
	Status             string    `json:"status"`
	LastSyncedAt       time.Time `json:"lastSyncedAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(languageID string) ([]File, error) {
	query := `
		SELECT id, project_id, language_id, path, checksum_original, checksum_translated, status, last_synced_at, updated_at
		FROM files WHERE language_id = $1 ORDER BY path ASC
	`
	rows, err := r.db.Query(query, languageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		var csOriginal, csTranslated sql.NullString
		if err := rows.Scan(
			&f.ID, &f.ProjectID, &f.LanguageID, &f.Path, &csOriginal, &csTranslated, &f.Status, &f.LastSyncedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		f.ChecksumOriginal = csOriginal.String
		f.ChecksumTranslated = csTranslated.String
		files = append(files, f)
	}
	return files, nil
}

func (r *Repository) GetByID(id string) (*File, error) {
	query := `
		SELECT id, project_id, language_id, path, checksum_original, checksum_translated, status, last_synced_at, updated_at
		FROM files WHERE id = $1
	`
	var f File
	var csOriginal, csTranslated sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&f.ID, &f.ProjectID, &f.LanguageID, &f.Path, &csOriginal, &csTranslated, &f.Status, &f.LastSyncedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	f.ChecksumOriginal = csOriginal.String
	f.ChecksumTranslated = csTranslated.String
	return &f, nil
}

func (r *Repository) Create(f *File) error {
	query := `
		INSERT INTO files (project_id, language_id, path, checksum_original, checksum_translated, status, last_synced_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, last_synced_at, updated_at
	`
	return r.db.QueryRow(
		query, f.ProjectID, f.LanguageID, f.Path, f.ChecksumOriginal, f.ChecksumTranslated, f.Status,
	).Scan(&f.ID, &f.LastSyncedAt, &f.UpdatedAt)
}

func (r *Repository) Update(id string, f *File) error {
	query := `
		UPDATE files
		SET path = $1, checksum_original = $2, checksum_translated = $3, status = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(
		query, f.Path, f.ChecksumOriginal, f.ChecksumTranslated, f.Status, id,
	).Scan(&f.UpdatedAt)
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM files WHERE id = $1", id)
	return err
}
