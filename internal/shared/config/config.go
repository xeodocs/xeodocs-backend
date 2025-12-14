package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	Environment    string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // Load .env file if it exists

	return &Config{
		Port:           getEnv("PORT", "12020"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/xeodocs?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "secret-key"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		AllowedOrigins: []string{"*"}, // TODO: Configure strictly for production
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
