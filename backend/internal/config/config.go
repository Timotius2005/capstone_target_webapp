package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	SecurityMode string
	Environment  string
}

func New() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		DatabaseURL: getEnv(
			"DATABASE_URL",
			"host=localhost user=fintech password=securepass dbname=nasabahdb port=5432 sslmode=disable",
		),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production-use-32-chars"),
		SecurityMode: getEnv("APP_SECURITY_MODE", "secure"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
