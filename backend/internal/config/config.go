package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration.
// DB fields are exposed separately so startup logs can show host/name without password.
type Config struct {
	Port string

	// Individual DB fields — used for logging and DSN construction.
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// DatabaseURL is the computed DSN. If DATABASE_URL is set in the environment
	// it takes priority over individual fields.
	DatabaseURL string

	JWTSecret    string
	SecurityMode string
	Environment  string
}

func New() *Config {
	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "fintech"),
		DBPassword:   getEnv("DB_PASSWORD", "securepass"),
		DBName:       getEnv("DB_NAME", "nasabahdb"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production-use-32-chars"),
		SecurityMode: getEnv("APP_SECURITY_MODE", "secure"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}

	// DATABASE_URL overrides individual fields if explicitly set.
	if url := os.Getenv("DATABASE_URL"); url != "" {
		cfg.DatabaseURL = url
	} else {
		cfg.DatabaseURL = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
		)
	}

	return cfg
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
