package main

import (
	"log"

	"github.com/xeodocs/xeodocs-backend/internal/scheduler"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

func main() {
	cfg := config.Load()

	// Start the scheduler
	scheduler.StartScheduler(cfg)

	log.Println("Scheduler service started")
}
