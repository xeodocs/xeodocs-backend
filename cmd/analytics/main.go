package main

import (
	"log"
	"net/http"

	"github.com/xeodocs/xeodocs-backend/internal/analytics"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()
	mux.HandleFunc("/track", analytics.TrackHandler(cfg))

	log.Printf("Starting Analytics Service on port %s", cfg.AnalyticsPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AnalyticsPort, mux))
}
