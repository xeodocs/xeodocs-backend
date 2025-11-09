package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/repository"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/clone-repo", repository.CloneRepoHandler(cfg))
	mux.HandleFunc("/internal/create-language-copies", repository.CreateLanguageCopiesHandler(cfg))
	mux.HandleFunc("/internal/sync-repo", repository.SyncRepoHandler(cfg))
	mux.HandleFunc("/internal/delete-repo", repository.DeleteRepoHandler(cfg))

	log.Printf("Starting Repository Service on port %s", cfg.RepositoryPort)
	log.Fatal(http.ListenAndServe(":"+cfg.RepositoryPort, mux))
}
