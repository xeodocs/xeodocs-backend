package config

import (
	"os"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	Port            string // for auth
	GatewayPort     string
	AuthServiceURL  string
}

func Load() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost/xeodocs_auth?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		Port:           getEnv("PORT", "8081"),
		GatewayPort:    getEnv("GATEWAY_PORT", "8080"),
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
