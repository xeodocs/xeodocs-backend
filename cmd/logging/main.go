package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/logging"
	"github.com/xeodocs/xeodocs-backend/internal/shared/auth"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/db"
)

func main() {
	cfg := config.Load()
	db.Init(cfg)
	defer db.Close()

	mux := http.NewServeMux()

	// POST /logs - create log (no auth, for other services)
	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			logging.CreateLogHandler(cfg)(w, r)
		} else if r.Method == http.MethodGet {
			// GET /logs - query logs (protected)
			auth.JWTMiddleware(cfg, "")(logging.QueryLogsHandler(cfg))(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Starting Logging Service on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
