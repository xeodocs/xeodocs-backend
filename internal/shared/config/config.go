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
	BuildPort            string
	BuildServiceURL      string
	AnalyticsPort        string
	AnalyticsServiceURL  string
	KafkaBrokers         string
	InfluxDBURL          string
	InfluxDBToken        string
	InfluxDBOrg          string
	InfluxDBBucket       string
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
		BuildPort:            getEnv("BUILD_PORT", "80"),
		BuildServiceURL:      getEnv("BUILD_SERVICE_URL", "http://localhost:80"),
		AnalyticsPort:        getEnv("ANALYTICS_PORT", "80"),
		AnalyticsServiceURL:  getEnv("ANALYTICS_SERVICE_URL", "http://localhost:80"),
		KafkaBrokers:         getEnv("KAFKA_BROKERS", "localhost:9092"),
		InfluxDBURL:          getEnv("INFLUXDB_URL", "http://localhost:8086"),
		InfluxDBToken:        getEnv("INFLUXDB_TOKEN", "my-super-secret-auth-token"),
		InfluxDBOrg:          getEnv("INFLUXDB_ORG", "xeodocs"),
		InfluxDBBucket:       getEnv("INFLUXDB_BUCKET", "analytics"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
