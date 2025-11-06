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
type BuildCommands []string

type Project struct {
	ID            int          `json:"id"`
	UserID        int          `json:"user_id"`
	Name          string       `json:"name"`
	RepoURL       string       `json:"repo_url"`
	Languages     Languages    `json:"languages"`
	BuildCommands BuildCommands `json:"build_commands"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
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

// Value implements driver.Valuer for JSONB
func (b BuildCommands) Value() (driver.Value, error) {
	return json.Marshal(b)
}

// Scan implements sql.Scanner for JSONB
func (b *BuildCommands) Scan(value interface{}) error {
	if value == nil {
		*b = BuildCommands{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

type CreateProjectRequest struct {
	Name          string       `json:"name"`
	RepoURL       string       `json:"repo_url"`
	Languages     Languages    `json:"languages"`
	BuildCommands BuildCommands `json:"build_commands"`
}

type UpdateProjectRequest struct {
	Name          *string      `json:"name,omitempty"`
	RepoURL       *string      `json:"repo_url,omitempty"`
	Languages     Languages    `json:"languages,omitempty"`
	BuildCommands BuildCommands `json:"build_commands,omitempty"`
}

func CreateProject(userID int, req CreateProjectRequest) (*Project, error) {
	project := &Project{
		UserID:        userID,
		Name:          req.Name,
		RepoURL:       req.RepoURL,
		Languages:     req.Languages,
		BuildCommands: req.BuildCommands,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `INSERT INTO projects (user_id, name, repo_url, languages, build_commands, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := db.DB.QueryRow(query, project.UserID, project.Name, project.RepoURL, project.Languages, project.BuildCommands, project.CreatedAt, project.UpdatedAt).Scan(&project.ID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func GetProjects(userID int) ([]Project, error) {
	query := `SELECT id, user_id, name, repo_url, languages, build_commands, created_at, updated_at FROM projects WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.RepoURL, &p.Languages, &p.BuildCommands, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func GetProjectByID(id int, userID int) (*Project, error) {
	project := &Project{}
	query := `SELECT id, user_id, name, repo_url, languages, build_commands, created_at, updated_at FROM projects WHERE id = $1 AND user_id = $2`
	row := db.DB.QueryRow(query, id, userID)
	err := row.Scan(&project.ID, &project.UserID, &project.Name, &project.RepoURL, &project.Languages, &project.BuildCommands, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return project, nil
}

func UpdateProject(id int, userID int, req UpdateProjectRequest) (*Project, error) {
	// First get the current project
	project, err := GetProjectByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.RepoURL != nil {
		project.RepoURL = *req.RepoURL
	}
	if req.Languages != nil {
		project.Languages = req.Languages
	}
	if req.BuildCommands != nil {
		project.BuildCommands = req.BuildCommands
	}
	project.UpdatedAt = time.Now()

	query := `UPDATE projects SET name = $1, repo_url = $2, languages = $3, build_commands = $4, updated_at = $5 WHERE id = $6 AND user_id = $7`
	_, err = db.DB.Exec(query, project.Name, project.RepoURL, project.Languages, project.BuildCommands, project.UpdatedAt, id, userID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func DeleteProject(id int, userID int) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`
	result, err := db.DB.Exec(query, id, userID)
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
