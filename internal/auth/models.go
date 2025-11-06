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
	RoleID    int       `json:"role_id"`
	Role      string    `json:"role"` // populated from join
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

type ChangePasswordRequest struct {
	Password string `json:"password"`
}

type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type RoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
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
	roleID := 3 // default viewer
	if req.Role != "" {
		role, err := GetRoleByName(req.Role)
		if err != nil {
			return nil, err
		}
		roleID = role.ID
	}

	user := &User{
		Username: req.Username,
		Password: req.Password,
		RoleID:   roleID,
	}

	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	query := `INSERT INTO users (username, password, role_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.DB.QueryRow(query, user.Username, user.Password, user.RoleID, time.Now()).Scan(&user.ID)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Now()
	// Populate Role name
	if role, err := GetRoleByID(user.RoleID); err == nil {
		user.Role = role.Name
	}
	return user, nil
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT u.id, u.username, u.password, u.role_id, r.name, u.created_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.username = $1`
	row := db.DB.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.RoleID, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func ChangePassword(userID int, newPassword string) error {
	user := &User{
		ID:       userID,
		Password: newPassword,
	}
	if err := user.HashPassword(); err != nil {
		return err
	}
	query := `UPDATE users SET password = $1 WHERE id = $2`
	_, err := db.DB.Exec(query, user.Password, userID)
	return err
}

// User CRUD functions

func GetUsers() ([]User, error) {
	query := `SELECT u.id, u.username, u.role_id, r.name, u.created_at FROM users u JOIN roles r ON u.role_id = r.id`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.RoleID, &u.Role, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT u.id, u.username, u.role_id, r.name, u.created_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.id = $1`
	row := db.DB.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.RoleID, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func UpdateUser(id int, req RegisterRequest) (*User, error) {
	user := &User{
		ID:       id,
		Username: req.Username,
	}

	if req.Role != "" {
		role, err := GetRoleByName(req.Role)
		if err != nil {
			return nil, err
		}
		user.RoleID = role.ID
		user.Role = role.Name
	}

	if req.Password != "" {
		user.Password = req.Password
		if err := user.HashPassword(); err != nil {
			return nil, err
		}
		query := `UPDATE users SET username = $1, password = $2, role_id = $3 WHERE id = $4`
		_, err := db.DB.Exec(query, user.Username, user.Password, user.RoleID, id)
		if err != nil {
			return nil, err
		}
	} else if req.Role != "" {
		query := `UPDATE users SET username = $1, role_id = $2 WHERE id = $3`
		_, err := db.DB.Exec(query, user.Username, user.RoleID, id)
		if err != nil {
			return nil, err
		}
	} else {
		query := `UPDATE users SET username = $1 WHERE id = $2`
		_, err := db.DB.Exec(query, user.Username, id)
		if err != nil {
			return nil, err
		}
	}

	return GetUserByID(id)
}

func DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.DB.Exec(query, id)
	return err
}

// Role CRUD functions

func CreateRole(req RoleRequest) (*Role, error) {
	role := &Role{
		Name:        req.Name,
		Description: req.Description,
	}

	query := `INSERT INTO roles (name, description, created_at) VALUES ($1, $2, $3) RETURNING id`
	err := db.DB.QueryRow(query, role.Name, role.Description, time.Now()).Scan(&role.ID)
	if err != nil {
		return nil, err
	}

	role.CreatedAt = time.Now()
	return role, nil
}

func GetRoles() ([]Role, error) {
	query := `SELECT id, name, description, created_at FROM roles`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, nil
}

func GetRoleByID(id int) (*Role, error) {
	role := &Role{}
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`
	row := db.DB.QueryRow(query, id)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return role, nil
}

func GetRoleByName(name string) (*Role, error) {
	role := &Role{}
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	row := db.DB.QueryRow(query, name)
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return role, nil
}

func UpdateRole(id int, req RoleRequest) (*Role, error) {
	query := `UPDATE roles SET name = $1, description = $2 WHERE id = $3`
	_, err := db.DB.Exec(query, req.Name, req.Description, id)
	if err != nil {
		return nil, err
	}
	return GetRoleByID(id)
}

func DeleteRole(id int) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := db.DB.Exec(query, id)
	return err
}
