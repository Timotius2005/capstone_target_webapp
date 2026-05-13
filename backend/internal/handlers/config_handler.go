package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"pt-dana-sejahtera/internal/middleware"
	"pt-dana-sejahtera/internal/security"
)

// ConfigHandler exposes GET/PUT /config/mode for runtime mode switching.
type ConfigHandler struct {
	log *zap.Logger
}

func NewConfigHandler(log *zap.Logger) *ConfigHandler {
	return &ConfigHandler{log: log}
}

// GetMode returns the current in-memory security mode.
// GET /config/mode — public (no auth required; frontend calls on every page load).
func (h *ConfigHandler) GetMode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"mode": security.GetMode(),
	})
}

// SetMode atomically changes the runtime security mode. Admin role required.
// PUT /config/mode
// Body: { "mode": "secure" | "sandbox" }
//
// OWASP A01 — restricted to admin role (enforced by RoleCheck middleware upstream).
// OWASP A08 — mode enum validated strictly; unknown values are rejected.
// OWASP A09 — every change is logged with actor identity and timestamp.
func (h *ConfigHandler) SetMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode field is required"})
		return
	}

	// OWASP A08: strict enum — reject anything outside the known set
	var newMode security.ModeValue
	switch req.Mode {
	case "secure":
		newMode = security.ModeSecure
	case "sandbox", "vulnerable": // accept legacy "vulnerable" alias
		newMode = security.ModeSandbox
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid mode: accepted values are 'secure' or 'sandbox'",
		})
		return
	}

	// Capture actor identity for audit log (injected by AuthRequired middleware)
	adminID, _ := c.Get(middleware.ContextUserID)
	adminUsername, _ := c.Get(middleware.ContextUsername)
	oldMode := security.GetMode()

	security.SetMode(newMode)

	// OWASP A09: structured audit log — timestamp + actor identity + old/new mode
	h.log.Info("Runtime security mode changed",
		zap.String("old_mode", oldMode),
		zap.String("new_mode", string(newMode)),
		zap.Any("admin_id", adminID),
		zap.Any("admin_username", adminUsername),
		zap.Time("changed_at", time.Now().UTC()),
	)

	c.JSON(http.StatusOK, gin.H{
		"mode":       security.GetMode(),
		"changed_by": adminUsername,
		"changed_at": time.Now().UTC(),
		"message":    "mode updated successfully",
	})
}
