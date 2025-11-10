package config

import (
	"os"
)

type Config struct {
	DatabaseURL          string
	JWTSecret            string
	Port                 string // for auth
	GatewayPort          string
	AuthServiceURL       string
	ProjectPort          string
	ProjectServiceURL    string
	RabbitMQURL          string
	RepositoryPort       string
	RepositoryServiceURL string
	LoggingServiceURL    string
}

func Load() *Config {
	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://user:password@localhost/xeodocs_db?sslmode=disable"),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
		Port:                 getEnv("PORT", "80"),
		GatewayPort:          getEnv("GATEWAY_PORT", "12020"),
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "http://localhost:80"),
		ProjectPort:          getEnv("PROJECT_PORT", "80"),
		ProjectServiceURL:    getEnv("PROJECT_SERVICE_URL", "http://localhost:80"),
		RabbitMQURL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RepositoryPort:       getEnv("REPOSITORY_PORT", "80"),
		RepositoryServiceURL: getEnv("REPOSITORY_SERVICE_URL", "http://localhost:80"),
		LoggingServiceURL:    getEnv("LOGGING_SERVICE_URL", "http://localhost:80"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
