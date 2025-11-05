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

	log.Printf("Starting Auth Service on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
