package languages

import (
	"database/sql"
	"time"
)

type Language struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(projectID string) ([]Language, error) {
	query := `
		SELECT id, project_id, code, name, domain, is_active, created_at, updated_at
		FROM languages WHERE project_id = $1 ORDER BY code ASC
	`
	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []Language
	for rows.Next() {
		var l Language
		var domain sql.NullString
		if err := rows.Scan(
			&l.ID, &l.ProjectID, &l.Code, &l.Name, &domain, &l.IsActive, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		l.Domain = domain.String
		languages = append(languages, l)
	}
	return languages, nil
}

func (r *Repository) GetByID(id string) (*Language, error) {
	query := `
		SELECT id, project_id, code, name, domain, is_active, created_at, updated_at
		FROM languages WHERE id = $1
	`
	var l Language
	var domain sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&l.ID, &l.ProjectID, &l.Code, &l.Name, &domain, &l.IsActive, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	l.Domain = domain.String
	return &l, nil
}

func (r *Repository) Create(l *Language) error {
	query := `
		INSERT INTO languages (project_id, code, name, domain, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query, l.ProjectID, l.Code, l.Name, l.Domain, l.IsActive,
	).Scan(&l.ID, &l.CreatedAt, &l.UpdatedAt)
}

func (r *Repository) Update(id string, l *Language) error {
	query := `
		UPDATE languages
		SET code = $1, name = $2, domain = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(
		query, l.Code, l.Name, l.Domain, l.IsActive, id,
	).Scan(&l.UpdatedAt)
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM languages WHERE id = $1", id)
	return err
}
