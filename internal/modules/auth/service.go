package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db     *sql.DB
	config *config.Config
}

func NewService(db *sql.DB, cfg *config.Config) *Service {
	return &Service{db: db, config: cfg}
}

type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	APIKey       string `json:"apiKey,omitempty"`
}

func (s *Service) Authenticate(email, password string) (*User, string, error) {
	var user User
	var apiKey sql.NullString
	err := s.db.QueryRow("SELECT id, name, email, password_hash, api_key FROM users WHERE email = $1", email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &apiKey,
	)
	if err == sql.ErrNoRows {
		return nil, "", errors.New("invalid credentials")
	}
	if err != nil {
		return nil, "", err
	}

	if apiKey.Valid {
		user.APIKey = apiKey.String
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return &user, token, nil
}

func (s *Service) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) ValidateAPIKey(apiKey string) (*User, error) {
	// Since API keys are hashed, we can't query directly.
	// We need to fetch all users with an API key and compare.
	// TODO: Optimization - Store a key ID or prefix to lookup efficiently.
	rows, err := s.db.Query("SELECT id, name, email, password_hash, api_key FROM users WHERE api_key IS NOT NULL AND api_key != ''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		var apiKeyHash string
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &apiKeyHash); err != nil {
			continue
		}
		user.APIKey = apiKeyHash

		if err := bcrypt.CompareHashAndPassword([]byte(apiKeyHash), []byte(apiKey)); err == nil {
			return &user, nil
		}
	}

	return nil, errors.New("invalid api key")
}
