package project

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
)

type Languages []string

type Project struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	DocURL         string    `json:"doc_url"`
	RepoURL        string    `json:"repo_url"`
	Languages      Languages `json:"languages"`
	BuildCommand   string    `json:"build_command"`
	ExportCommand  string    `json:"export_command"`
	PreviewCommand string    `json:"preview_command"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Value implements driver.Valuer for JSONB
func (l Languages) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// Scan implements sql.Scanner for JSONB
func (l *Languages) Scan(value interface{}) error {
	if value == nil {
		*l = Languages{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, l)
}

type CreateProjectRequest struct {
	Name           string    `json:"name"`
	DocURL         string    `json:"doc_url"`
	RepoURL        string    `json:"repo_url"`
	Languages      Languages `json:"languages"`
	BuildCommand   string    `json:"build_command"`
	ExportCommand  string    `json:"export_command"`
	PreviewCommand string    `json:"preview_command"`
}

type UpdateProjectRequest struct {
	Name           *string   `json:"name,omitempty"`
	DocURL         *string   `json:"doc_url,omitempty"`
	RepoURL        *string   `json:"repo_url,omitempty"`
	Languages      Languages `json:"languages,omitempty"`
	BuildCommand   *string   `json:"build_command,omitempty"`
	ExportCommand  *string   `json:"export_command,omitempty"`
	PreviewCommand *string   `json:"preview_command,omitempty"`
}

func CreateProject(req CreateProjectRequest) (*Project, error) {
	project := &Project{
		Name:           req.Name,
		DocURL:         req.DocURL,
		RepoURL:        req.RepoURL,
		Languages:      req.Languages,
		BuildCommand:   req.BuildCommand,
		ExportCommand:  req.ExportCommand,
		PreviewCommand: req.PreviewCommand,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	query := `INSERT INTO projects (name, doc_url, repo_url, languages, build_command, export_command, preview_command, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err := db.DB.QueryRow(query, project.Name, project.DocURL, project.RepoURL, project.Languages, project.BuildCommand, project.ExportCommand, project.PreviewCommand, project.CreatedAt, project.UpdatedAt).Scan(&project.ID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func GetProjects() ([]Project, error) {
	query := `SELECT id, name, doc_url, repo_url, languages, build_command, export_command, preview_command, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.Name, &p.DocURL, &p.RepoURL, &p.Languages, &p.BuildCommand, &p.ExportCommand, &p.PreviewCommand, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func GetProjectByID(id int) (*Project, error) {
	project := &Project{}
	query := `SELECT id, name, doc_url, repo_url, languages, build_command, export_command, preview_command, created_at, updated_at FROM projects WHERE id = $1`
	row := db.DB.QueryRow(query, id)
	err := row.Scan(&project.ID, &project.Name, &project.DocURL, &project.RepoURL, &project.Languages, &project.BuildCommand, &project.ExportCommand, &project.PreviewCommand, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return project, nil
}

func UpdateProject(id int, req UpdateProjectRequest) (*Project, error) {
	// First get the current project
	project, err := GetProjectByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.DocURL != nil {
		project.DocURL = *req.DocURL
	}
	if req.RepoURL != nil {
		project.RepoURL = *req.RepoURL
	}
	if req.Languages != nil {
		project.Languages = req.Languages
	}
	if req.BuildCommand != nil {
		project.BuildCommand = *req.BuildCommand
	}
	if req.ExportCommand != nil {
		project.ExportCommand = *req.ExportCommand
	}
	if req.PreviewCommand != nil {
		project.PreviewCommand = *req.PreviewCommand
	}
	project.UpdatedAt = time.Now()

	query := `UPDATE projects SET name = $1, doc_url = $2, repo_url = $3, languages = $4, build_command = $5, export_command = $6, preview_command = $7, updated_at = $8 WHERE id = $9`
	_, err = db.DB.Exec(query, project.Name, project.DocURL, project.RepoURL, project.Languages, project.BuildCommand, project.ExportCommand, project.PreviewCommand, project.UpdatedAt, id)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func DeleteProject(id int) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := db.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("project not found")
	}

	return nil
}
