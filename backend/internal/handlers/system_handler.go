package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
)

// SystemHandler serves the public mode-switch and vuln-config APIs.
// No authentication is required — protected only by optional LabKeyRequired middleware.
type SystemHandler struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewSystemHandler(db *gorm.DB, log *zap.Logger) *SystemHandler {
	return &SystemHandler{db: db, log: log}
}

// GetMode returns the current global security mode.
// GET /api/system/mode — public, no auth required.
func (h *SystemHandler) GetMode(c *gin.Context) {
	mode := toExternalMode(security.GetMode())
	c.JSON(http.StatusOK, gin.H{
		"mode":      mode,
		"timestamp": time.Now().UTC(),
	})
}

// SetMode switches the global security mode, persists it to the database,
// and records an audit log entry.
// PUT /api/system/mode — public (optionally protected by LabKeyRequired middleware).
//
// Body: { "mode": "secure" | "vulnerable" }
func (h *SystemHandler) SetMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode field is required"})
		return
	}

	var newMode security.ModeValue
	switch req.Mode {
	case "secure":
		newMode = security.ModeSecure
	case "vulnerable", "sandbox":
		newMode = security.ModeSandbox
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid mode: accepted values are 'secure' or 'vulnerable'",
		})
		return
	}

	oldModeDisplay := toExternalMode(security.GetMode())
	newModeDisplay := toExternalMode(string(newMode))
	now := time.Now().UTC()

	// Update in-memory mode (atomic).
	security.SetMode(newMode)

	// When switching back to secure, reset vuln config so that the next
	// vulnerable-mode session starts with all categories enabled by default.
	if newMode == security.ModeSecure {
		security.ResetVulnConfig()
	}

	// Persist to DB — always update the singleton row (no-op if db is nil, e.g. in tests).
	if h.db != nil {
		if err := h.db.Model(&models.SystemSetting{}).
			Where("id = ?", models.SystemSettingsID).
			Updates(map[string]interface{}{
				"mode":       newModeDisplay,
				"updated_at": now,
			}).Error; err != nil {
			h.log.Warn("Failed to persist mode change to DB", zap.Error(err))
		}

		// Append immutable audit log entry.
		if err := h.db.Create(&models.ModeChangeLog{
			PreviousMode: oldModeDisplay,
			NewMode:      newModeDisplay,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			ChangedAt:    now,
		}).Error; err != nil {
			h.log.Warn("Failed to write mode_change_logs entry", zap.Error(err))
		}
	}

	h.log.Info("System mode changed via public API",
		zap.String("old_mode", oldModeDisplay),
		zap.String("new_mode", newModeDisplay),
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.Time("changed_at", now),
	)

	c.JSON(http.StatusOK, gin.H{
		"mode":       newModeDisplay,
		"changed_at": now,
		"message":    "mode updated successfully",
	})
}

// GetVulnConfig returns the current per-category vulnerability configuration.
// GET /api/system/vuln-config — public, no auth required.
// Returns the config regardless of mode so the frontend can always read it.
func (h *SystemHandler) GetVulnConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"config":    security.GetVulnConfig(),
		"mode":      toExternalMode(security.GetMode()),
		"timestamp": time.Now().UTC(),
	})
}

// SetVulnConfig updates the per-category vulnerability configuration.
// PUT /api/system/vuln-config — public (optionally protected by LabKeyRequired).
//
// Only meaningful when mode = "vulnerable"; returns 400 in secure mode
// to prevent confusion (secure mode ignores the config anyway).
//
// Body: full VulnConfig JSON object.
func (h *SystemHandler) SetVulnConfig(c *gin.Context) {
	if !security.IsVulnerable() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "vulnerability config can only be modified in vulnerable mode; switch mode first",
		})
		return
	}

	var cfg security.VulnConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config: " + err.Error()})
		return
	}

	security.SetVulnConfig(cfg)

	h.log.Info("Vulnerability config updated",
		zap.String("ip", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	c.JSON(http.StatusOK, gin.H{
		"config":  security.GetVulnConfig(),
		"message": "vulnerability config updated successfully",
	})
}

// toExternalMode translates the internal "sandbox" alias to the public-facing
// "vulnerable" label used in the API contract.
func toExternalMode(m string) string {
	if m == "sandbox" {
		return "vulnerable"
	}
	return m
}
