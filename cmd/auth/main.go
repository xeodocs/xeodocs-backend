package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/auth"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
)

func main() {
	cfg := config.Load()
	db.Init(cfg)
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", auth.JWTMiddleware(cfg, "admin")(auth.RegisterHandler(cfg)))
	mux.HandleFunc("/auth/login", auth.LoginHandler(cfg))
	mux.HandleFunc("/auth/change-password", auth.JWTMiddleware(cfg, "")(auth.ChangePasswordHandler(cfg)))

	// Users CRUD - protected
	mux.HandleFunc("/users", auth.JWTMiddleware(cfg, "admin")(auth.ListUsersHandler(cfg)))
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 7 && path[:7] == "/users/" {
			id := path[7:]
			if id == "" {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				auth.JWTMiddleware(cfg, "admin")(auth.GetUserHandler(cfg))(w, r)
			case http.MethodPut:
				auth.JWTMiddleware(cfg, "admin")(auth.UpdateUserHandler(cfg))(w, r)
			case http.MethodDelete:
				auth.JWTMiddleware(cfg, "admin")(auth.DeleteUserHandler(cfg))(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Roles CRUD - protected
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth.JWTMiddleware(cfg, "admin")(auth.ListRolesHandler(cfg))(w, r)
		case http.MethodPost:
			auth.JWTMiddleware(cfg, "admin")(auth.CreateRoleHandler(cfg))(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/roles/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 7 && path[:7] == "/roles/" {
			id := path[7:]
			if id == "" {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				auth.JWTMiddleware(cfg, "admin")(auth.GetRoleHandler(cfg))(w, r)
			case http.MethodPut:
				auth.JWTMiddleware(cfg, "admin")(auth.UpdateRoleHandler(cfg))(w, r)
			case http.MethodDelete:
				auth.JWTMiddleware(cfg, "admin")(auth.DeleteRoleHandler(cfg))(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	log.Printf("Starting Auth Service on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
