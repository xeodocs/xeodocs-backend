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
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) GetByID(id string) (*User, error) {
	var u User
	err := r.db.QueryRow("SELECT id, name, email, api_key, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) Create(name, email, passwordHash, apiKey string) (*User, error) {
	var u User
	err := r.db.QueryRow(`
		INSERT INTO users (name, email, password_hash, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, name, email, api_key, created_at, updated_at
	`, name, email, passwordHash, apiKey).Scan(&u.ID, &u.Name, &u.Email, &u.APIKey, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
