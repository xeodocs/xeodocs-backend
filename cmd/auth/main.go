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
	mux.HandleFunc("/auth/register", auth.RegisterHandler(cfg))
	mux.HandleFunc("/auth/login", auth.LoginHandler(cfg))

	// Users CRUD
	mux.HandleFunc("/users", auth.ListUsersHandler(cfg))
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
				auth.GetUserHandler(cfg)(w, r)
			case http.MethodPut:
				auth.UpdateUserHandler(cfg)(w, r)
			case http.MethodDelete:
				auth.DeleteUserHandler(cfg)(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	// Roles CRUD
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth.ListRolesHandler(cfg)(w, r)
		case http.MethodPost:
			auth.CreateRoleHandler(cfg)(w, r)
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
				auth.GetRoleHandler(cfg)(w, r)
			case http.MethodPut:
				auth.UpdateRoleHandler(cfg)(w, r)
			case http.MethodDelete:
				auth.DeleteRoleHandler(cfg)(w, r)
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
