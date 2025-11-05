package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func CreateUser(req RegisterRequest) (*User, error) {
	if req.Role == "" {
		req.Role = "viewer" // default role
	}

	user := &User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	query := `INSERT INTO users (username, password, role, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.DB.QueryRow(query, user.Username, user.Password, user.Role, time.Now()).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Now()
	return user, nil
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, role, created_at FROM users WHERE username = $1`
	row := db.DB.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
