package security

import (
	"sync"

	"go.uber.org/zap"
)

// ModeValue is the canonical type for the runtime security mode.
type ModeValue string

const (
	ModeSecure  ModeValue = "secure"
	ModeSandbox ModeValue = "sandbox"
)

var (
	mu          sync.RWMutex
	currentMode ModeValue = ModeSecure // safe default; overridden by Init()
)

// Init sets the starting runtime mode from the value loaded in config at startup.
// Call once before serving requests. Accepts "secure", "sandbox", or legacy "vulnerable".
func Init(modeStr string) {
	switch modeStr {
	case "sandbox", "vulnerable":
		SetMode(ModeSandbox)
	default:
		SetMode(ModeSecure)
	}
}

// SetMode atomically changes the active security mode at runtime without a restart.
func SetMode(mode ModeValue) {
	mu.Lock()
	defer mu.Unlock()
	currentMode = mode
}

// GetMode returns the current mode string ("secure" or "sandbox").
func GetMode() string {
	mu.RLock()
	defer mu.RUnlock()
	return string(currentMode)
}

// IsSecure returns true when the runtime mode is "secure".
func IsSecure() bool {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode == ModeSecure
}

// IsVulnerable returns true when the runtime mode is "sandbox".
// Preserved for compatibility with existing OWASP injection-point code.
func IsVulnerable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode == ModeSandbox
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
			zap.String("note", "OWASP sandbox/demo mode ACTIVE — controlled simulation only"),
		)
	}
}
