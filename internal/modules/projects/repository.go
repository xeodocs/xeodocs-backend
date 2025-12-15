package projects

import (
	"database/sql"
	"time"
)

type Project struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Slug             string       `json:"slug"`
	SourceWebsiteURL string       `json:"sourceWebsiteUrl,omitempty"`
	SourceRepoURL    string       `json:"sourceRepoUrl"`
	SourceBranch     string       `json:"sourceBranch"`
	Description      string       `json:"description,omitempty"`
	IsActive         bool         `json:"isActive"`
	LastRepoCheckAt  sql.NullTime `json:"lastRepoCheckAt,omitempty"`
	LastCommitHash   string       `json:"lastCommitHash,omitempty"`
	CreatedAt        time.Time    `json:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]Project, error) {
	query := `
		SELECT id, name, slug, source_website_url, source_repo_url, source_branch, description, is_active, last_repo_check_at, last_commit_hash, created_at, updated_at
		FROM projects ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var desc, websiteURL, commitHash sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &websiteURL, &p.SourceRepoURL, &p.SourceBranch, &desc, &p.IsActive, &p.LastRepoCheckAt, &commitHash, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.Description = desc.String
		p.SourceWebsiteURL = websiteURL.String
		p.LastCommitHash = commitHash.String
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *Repository) GetByID(id string) (*Project, error) {
	query := `
		SELECT id, name, slug, source_website_url, source_repo_url, source_branch, description, is_active, last_repo_check_at, last_commit_hash, created_at, updated_at
		FROM projects WHERE id = $1
	`
	var p Project
	var desc, websiteURL, commitHash sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &websiteURL, &p.SourceRepoURL, &p.SourceBranch, &desc, &p.IsActive, &p.LastRepoCheckAt, &commitHash, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Description = desc.String
	p.SourceWebsiteURL = websiteURL.String
	p.LastCommitHash = commitHash.String
	return &p, nil
}

func (r *Repository) GetActive() ([]Project, error) {
	query := `
		SELECT id, name, slug, source_website_url, source_repo_url, source_branch, description, is_active, last_repo_check_at, last_commit_hash, created_at, updated_at
		FROM projects WHERE is_active = true ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		var desc, websiteURL, commitHash sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &websiteURL, &p.SourceRepoURL, &p.SourceBranch, &desc, &p.IsActive, &p.LastRepoCheckAt, &commitHash, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.Description = desc.String
		p.SourceWebsiteURL = websiteURL.String
		p.LastCommitHash = commitHash.String
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *Repository) Create(p *Project) error {
	query := `
		INSERT INTO projects (name, slug, source_website_url, source_repo_url, source_branch, description, is_active, last_repo_check_at, last_commit_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query, p.Name, p.Slug, p.SourceWebsiteURL, p.SourceRepoURL, p.SourceBranch, p.Description, p.IsActive, p.LastRepoCheckAt, p.LastCommitHash,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) Update(id string, p *Project) error {
	query := `
		UPDATE projects
		SET name = $1, slug = $2, source_website_url = $3, source_repo_url = $4, source_branch = $5, description = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`
	// Note: We don't update sync fields here usually, but if we wanted to we could.
	// For now, keeping original Update logic roughly same, but sync fields are updated separately or we should include them?
	// The Handler uses this Update, and doesn't set Sync fields. So it's fine.
	return r.db.QueryRow(
		query, p.Name, p.Slug, p.SourceWebsiteURL, p.SourceRepoURL, p.SourceBranch, p.Description, p.IsActive, id,
	).Scan(&p.UpdatedAt)
}

func (r *Repository) UpdateSyncStatus(id string, lastCheckAt time.Time, lastCommitHash string) error {
	query := `
		UPDATE projects
		SET last_repo_check_at = $1, last_commit_hash = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, lastCheckAt, lastCommitHash, id)
	return err
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM projects WHERE id = $1", id)
	return err
}
