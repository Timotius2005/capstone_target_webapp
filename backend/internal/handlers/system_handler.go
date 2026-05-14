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

// SystemHandler serves the public mode-switch API at /api/system/mode.
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

	// Persist to DB — always update the singleton row.
	if err := h.db.Model(&models.SystemSetting{}).
		Where("id = ?", models.SystemSettingsID).
		Updates(map[string]interface{}{
			"mode":       newModeDisplay,
			"updated_at": now,
		}).Error; err != nil {
		h.log.Warn("Failed to persist mode change to DB", zap.Error(err))
		// Non-fatal: in-memory mode is already updated, system still works.
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

// toExternalMode translates the internal "sandbox" alias to the public-facing
// "vulnerable" label used in the API contract.
func toExternalMode(m string) string {
	if m == "sandbox" {
		return "vulnerable"
	}
	return m
}
