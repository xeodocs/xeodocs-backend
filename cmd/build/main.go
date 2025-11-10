package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/build"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/internal/build", build.BuildHandler(cfg))
	mux.HandleFunc("/internal/export", build.ExportHandler(cfg))
	mux.HandleFunc("/internal/preview", build.PreviewHandler(cfg))

	log.Printf("Starting Build Service on port %s", cfg.BuildPort)
	log.Fatal(http.ListenAndServe(":"+cfg.BuildPort, mux))
}
