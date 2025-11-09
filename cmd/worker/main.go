package main

import (
	"log"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/worker"
)

func main() {
	cfg := config.Load()

	// Initialize worker
	worker.Start(cfg)

	log.Println("Worker service started")
	select {} // Keep running
}
