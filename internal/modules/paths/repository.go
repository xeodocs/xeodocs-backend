package paths

import (
	"database/sql"
)

type IgnoredPath struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Pattern   string `json:"pattern"`
}

type SpecialPath struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Purpose   string `json:"purpose"`
	Pattern   string `json:"pattern"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Ignored Paths

func (r *Repository) GetIgnoredPaths(projectID string) ([]IgnoredPath, error) {
	rows, err := r.db.Query("SELECT id, project_id, pattern FROM ignored_paths WHERE project_id = $1", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []IgnoredPath
	for rows.Next() {
		var p IgnoredPath
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Pattern); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *Repository) CreateIgnoredPath(p *IgnoredPath) error {
	return r.db.QueryRow(
		"INSERT INTO ignored_paths (project_id, pattern) VALUES ($1, $2) RETURNING id",
		p.ProjectID, p.Pattern,
	).Scan(&p.ID)
}

func (r *Repository) DeleteIgnoredPath(id string) error {
	_, err := r.db.Exec("DELETE FROM ignored_paths WHERE id = $1", id)
	return err
}

// Special Paths

func (r *Repository) GetSpecialPaths(projectID string) ([]SpecialPath, error) {
	rows, err := r.db.Query("SELECT id, project_id, purpose, pattern FROM special_paths WHERE project_id = $1", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []SpecialPath
	for rows.Next() {
		var p SpecialPath
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Purpose, &p.Pattern); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

func (r *Repository) CreateSpecialPath(p *SpecialPath) error {
	return r.db.QueryRow(
		"INSERT INTO special_paths (project_id, purpose, pattern) VALUES ($1, $2, $3) RETURNING id",
		p.ProjectID, p.Purpose, p.Pattern,
	).Scan(&p.ID)
}

func (r *Repository) DeleteSpecialPath(id string) error {
	_, err := r.db.Exec("DELETE FROM special_paths WHERE id = $1", id)
	return err
}
