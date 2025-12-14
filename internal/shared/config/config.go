package config

import (
	"fmt"
	"os"

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

	port, err := getRequiredEnv("PORT")
	if err != nil {
		return nil, err
	}

	databaseURL, err := getRequiredEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	jwtSecret, err := getRequiredEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	environment, err := getRequiredEnv("ENVIRONMENT")
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:           port,
		DatabaseURL:    databaseURL,
		JWTSecret:      jwtSecret,
		Environment:    environment,
		AllowedOrigins: []string{"*"}, // TODO: Configure strictly for production
	}, nil
}

func getRequiredEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("missing required environment variable: %s", key)
}
