package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/auth"
	"github.com/xeodocs/xeodocs-backend/internal/project"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
)

func main() {
	cfg := config.Load()
	db.Init(cfg)
	defer db.Close()

	mux := http.NewServeMux()

	// Projects CRUD - protected
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth.JWTMiddleware(cfg, "")(project.ListProjectsHandler(cfg))(w, r)
		case http.MethodPost:
			auth.JWTMiddleware(cfg, "")(project.CreateProjectHandler(cfg))(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 10 && path[:10] == "/projects/" {
			id := path[10:]
			if id == "" {
				http.NotFound(w, r)
				return
			}
			switch r.Method {
			case http.MethodGet:
				auth.JWTMiddleware(cfg, "")(project.GetProjectHandler(cfg))(w, r)
			case http.MethodPut:
				auth.JWTMiddleware(cfg, "")(project.UpdateProjectHandler(cfg))(w, r)
			case http.MethodDelete:
				auth.JWTMiddleware(cfg, "")(project.DeleteProjectHandler(cfg))(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	log.Printf("Starting Project Service on port %s", cfg.ProjectPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ProjectPort, mux))
}
