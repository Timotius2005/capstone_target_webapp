package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a zap logger. Uses production config in secure mode, debug in vulnerable.
func New() *zap.Logger {
	mode := os.Getenv("APP_SECURITY_MODE")

	var cfg zap.Config
	if mode == "vulnerable" {
		// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
		// Debug mode: verbose logging including sensitive request data
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		// Suppress caller info in production to reduce fingerprinting
		cfg.DisableCaller = true
	}

	log, err := cfg.Build()
	if err != nil {
		// Fallback to no-op logger if build fails
		return zap.NewNop()
	}
	return log
}
