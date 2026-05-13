package security

import (
	"os"

	"go.uber.org/zap"
)

// IsSecure returns true when APP_SECURITY_MODE=secure.
func IsSecure() bool {
	return os.Getenv("APP_SECURITY_MODE") == "secure"
}

// IsVulnerable returns true when APP_SECURITY_MODE=vulnerable (or unset).
func IsVulnerable() bool {
	return !IsSecure()
}

// GetMode returns the current security mode string.
func GetMode() string {
	if IsSecure() {
		return "secure"
	}
	return "vulnerable"
}

// LogMode prints the active security mode at startup.
func LogMode(log *zap.Logger) {
	mode := GetMode()
	if IsSecure() {
		log.Info("Security mode active",
			zap.String("mode", mode),
			zap.String("note", "All OWASP protections ENABLED"),
		)
	} else {
		log.Warn("Security mode active",
			zap.String("mode", mode),
			zap.String("note", "OWASP vulnerabilities INTENTIONALLY ENABLED — demo only"),
		)
	}
}
