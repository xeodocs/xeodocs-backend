package users

import (
	"database/sql"
	"fmt"
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

func (r *Repository) Update(id string, name, email, passwordHash, apiKeyHash string) (*User, error) {
	var u User
	var apiKey sql.NullString

	query := "UPDATE users SET name = $1, email = $2, updated_at = NOW()"
	args := []interface{}{name, email}
	argID := 3

	if passwordHash != "" {
		query += ", password_hash = $" + string(rune(argID+'0')) // simplified for single digit, but let's be safe
		// actually, let's use fmt.Sprintf or just append
		query += fmt.Sprintf(", password_hash = $%d", argID)
		args = append(args, passwordHash)
		argID++
	}

	if apiKeyHash != "" {
		query += fmt.Sprintf(", api_key = $%d", argID)
		args = append(args, apiKeyHash)
		argID++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, name, email, api_key, created_at, updated_at", argID)
	args = append(args, id)

	err := r.db.QueryRow(query, args...).Scan(&u.ID, &u.Name, &u.Email, &apiKey, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return nil, err
	}
	if apiKey.Valid {
		u.APIKey = apiKey.String
	}
	return &u, nil
}

func (r *Repository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
