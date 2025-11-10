package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xeodocs/xeodocs-backend/internal/shared/auth"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/logging"
)

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) *int {
	claims, ok := ctx.Value("claims").(*auth.Claims)
	if !ok {
		return nil
	}
	return &claims.UserID
}

func JWTMiddleware(cfg *config.Config, requiredRole string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateJWT(tokenString, cfg)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if requiredRole != "" && claims.RoleID != 1 {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), "claims", claims)
			r = r.WithContext(ctx)
			next(w, r)
		}
	}
}

func RegisterHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := CreateUser(req)
		if err != nil {
			log.Println("Error creating user:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		token, err := auth.GenerateJWT(user.ID, user.Username, user.RoleID, cfg)
		if err != nil {
			log.Println("Error generating token:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the user registration
		message := "User registered: " + user.Username
		logging.LogActivity(cfg.LoggingServiceURL, "user_registered", message, &user.ID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := GetUserByUsername(req.Username)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !user.CheckPassword(req.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := auth.GenerateJWT(user.ID, user.Username, user.RoleID, cfg)
		if err != nil {
			log.Println("Error generating token:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the user login
		message := "User logged in: " + user.Username
		logging.LogActivity(cfg.LoggingServiceURL, "user_login", message, &user.ID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func ChangePasswordHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		claims := r.Context().Value("claims").(*auth.Claims)

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := ChangePassword(claims.UserID, req.Password); err != nil {
			log.Println("Error changing password:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the password change
		userID := getUserIDFromContext(r.Context())
		message := "Password changed for user ID: " + strconv.Itoa(claims.UserID)
		logging.LogActivity(cfg.LoggingServiceURL, "password_changed", message, userID, nil, "info")

		w.WriteHeader(http.StatusNoContent)
	}
}

// User CRUD handlers

func ListUsersHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		users, err := GetUsers()
		if err != nil {
			log.Println("Error getting users:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the user listing
		userID := getUserIDFromContext(r.Context())
		message := "Users listed"
		logging.LogActivity(cfg.LoggingServiceURL, "users_listed", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func GetUserHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/users/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		user, err := GetUserByID(id)
		if err != nil {
			if err.Error() == "user not found" {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				log.Println("Error getting user:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Log the user retrieval
		userID := getUserIDFromContext(r.Context())
		message := "User retrieved: " + user.Username
		logging.LogActivity(cfg.LoggingServiceURL, "user_retrieved", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func UpdateUserHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/users/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		user, err := UpdateUser(id, req)
		if err != nil {
			log.Println("Error updating user:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the user update
		userID := getUserIDFromContext(r.Context())
		message := "User updated: " + user.Username
		logging.LogActivity(cfg.LoggingServiceURL, "user_updated", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func DeleteUserHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/users/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		err = DeleteUser(id)
		if err != nil {
			log.Println("Error deleting user:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the user deletion
		userID := getUserIDFromContext(r.Context())
		message := "User deleted with ID: " + strconv.Itoa(id)
		logging.LogActivity(cfg.LoggingServiceURL, "user_deleted", message, userID, nil, "info")

		w.WriteHeader(http.StatusNoContent)
	}
}

// Role CRUD handlers

func CreateRoleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		role, err := CreateRole(req)
		if err != nil {
			log.Println("Error creating role:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the role creation
		userID := getUserIDFromContext(r.Context())
		message := "Role created: " + role.Name
		logging.LogActivity(cfg.LoggingServiceURL, "role_created", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(role)
	}
}

func ListRolesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roles, err := GetRoles()
		if err != nil {
			log.Println("Error getting roles:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the role listing
		userID := getUserIDFromContext(r.Context())
		message := "Roles listed"
		logging.LogActivity(cfg.LoggingServiceURL, "roles_listed", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roles)
	}
}

func GetRoleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/roles/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid role ID", http.StatusBadRequest)
			return
		}

		role, err := GetRoleByID(id)
		if err != nil {
			if err.Error() == "role not found" {
				http.Error(w, "Role not found", http.StatusNotFound)
			} else {
				log.Println("Error getting role:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Log the role retrieval
		userID := getUserIDFromContext(r.Context())
		message := "Role retrieved: " + role.Name
		logging.LogActivity(cfg.LoggingServiceURL, "role_retrieved", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(role)
	}
}

func UpdateRoleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/roles/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid role ID", http.StatusBadRequest)
			return
		}

		var req RoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		role, err := UpdateRole(id, req)
		if err != nil {
			log.Println("Error updating role:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the role update
		userID := getUserIDFromContext(r.Context())
		message := "Role updated: " + role.Name
		logging.LogActivity(cfg.LoggingServiceURL, "role_updated", message, userID, nil, "info")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(role)
	}
}

func DeleteRoleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := r.URL.Path[len("/roles/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid role ID", http.StatusBadRequest)
			return
		}

		err = DeleteRole(id)
		if err != nil {
			log.Println("Error deleting role:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Log the role deletion
		userID := getUserIDFromContext(r.Context())
		message := "Role deleted with ID: " + strconv.Itoa(id)
		logging.LogActivity(cfg.LoggingServiceURL, "role_deleted", message, userID, nil, "info")

		w.WriteHeader(http.StatusNoContent)
	}
}
