package users

import (
	"database/sql"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	APIKey    string    `json:"apiKey,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]User, error) {
	rows, err := r.db.Query("SELECT id, name, email, api_key, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var apiKey sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &apiKey, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		if apiKey.Valid {
			u.APIKey = apiKey.String
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) GetByID(id string) (*User, error) {
	var u User
	var apiKey sql.NullString
	err := r.db.QueryRow("SELECT id, name, email, api_key, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Name, &u.Email, &apiKey, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if apiKey.Valid {
		u.APIKey = apiKey.String
	}
	return &u, nil
}

func (r *Repository) Create(name, email, passwordHash, apiKey string) (*User, error) {
	var u User
	var apiKeyResult sql.NullString
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, email, api_key, created_at, updated_at
	`, name, email, passwordHash, apiKey).Scan(&u.ID, &u.Name, &u.Email, &apiKeyResult, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if apiKeyResult.Valid {
		u.APIKey = apiKeyResult.String
	}
	return &u, nil
}

func (r *Repository) Update(id string, name, email string) (*User, error) {
	var u User
	var apiKey sql.NullString
	err := r.db.QueryRow(`
		UPDATE users
		SET name = $1, email = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, email, api_key, created_at, updated_at
	`, name, email, id).Scan(&u.ID, &u.Name, &u.Email, &apiKey, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if apiKey.Valid {
		u.APIKey = apiKey.String
	}
	return &u, nil
}
