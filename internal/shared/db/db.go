package db

import (
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) {
	var err error
	DB, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := goose.Up(DB, "/app/internal/shared/db/migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Database connected")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
