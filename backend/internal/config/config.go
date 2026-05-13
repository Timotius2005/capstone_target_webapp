package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration.
// All sensitive and infrastructure values must be supplied via environment
// variables or a .env file. No defaults are provided for those fields.
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
		Port:         getEnv("PORT", ""),
		DBHost:       getEnv("DB_HOST", ""),
		DBPort:       getEnv("DB_PORT", ""),
		DBUser:       getEnv("DB_USER", ""),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", ""),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		SecurityMode: getEnv("APP_SECURITY_MODE", "secure"),
		Environment:  getEnv("ENVIRONMENT", "production"),
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
