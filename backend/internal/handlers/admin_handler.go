package handlers

import (
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/security"
)

type AdminHandler struct {
	userRepo repository.UserRepository
	log      *zap.Logger
}

func NewAdminHandler(userRepo repository.UserRepository, log *zap.Logger) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, log: log}
}

// ListUsers godoc
// GET /api/v1/admin/users
// OWASP API5: admin only (secure); exposed without check (vulnerable v0).
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := h.userRepo.ListAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	result := make([]interface{}, len(users))
	for i, u := range users {
		if security.IsVulnerable() {
			// TODO: Vulnerability Injection Point — OWASP API3 (BOPLA)
			// Returns password_hash, login_attempts, internal fields
			result[i] = u.ToVulnerableResponse()
		} else {
			result[i] = u.ToResponse()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateRole godoc
// PUT /api/v1/admin/users/:id/role
// OWASP API3: admin can change roles; secure mode validates transition.
func (h *AdminHandler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var body struct {
		Role string `json:"role" binding:"required,oneof=admin staff nasabah"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	oldRole := user.Role
	user.Role = body.Role

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	h.log.Info("User role updated",
		zap.String("user_id", id.String()),
		zap.String("old_role", oldRole),
		zap.String("new_role", body.Role),
		zap.String("updated_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "role updated",
		"user":    user.ToResponse(),
	})
}

// Stats godoc
// GET /api/v1/admin/stats
func (h *AdminHandler) Stats(c *gin.Context) {
	// Basic stats from user count
	users, total, _ := h.userRepo.ListAll(1, 1)
	_ = users

	c.JSON(http.StatusOK, gin.H{
		"total_users":   total,
		"security_mode": security.GetMode(),
		"version":       "v1",
	})
}

// ListUsersPublic godoc — v0 deprecated endpoint.
// TODO: Vulnerability Injection Point — OWASP API9 (Improper Inventory Management)
// Hidden endpoint exposed only in vulnerable mode, no auth required.
func (h *AdminHandler) ListUsersPublic(c *gin.Context) {
	h.log.Warn("[VULNERABLE] v0 public user list accessed — no authentication required")

	users, total, err := h.userRepo.ListAll(1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}

	result := make([]models.UserVulnerableResponse, len(users))
	for i, u := range users {
		result[i] = u.ToVulnerableResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"note":  "[VULNERABLE] Unauthenticated endpoint — OWASP API9",
	})
}

// Debug godoc — hidden debug endpoint, vulnerable mode only.
// TODO: Vulnerability Injection Point — OWASP API8 (Security Misconfiguration)
// TODO: Vulnerability Injection Point — OWASP API9 (Improper Inventory Management)
func (h *AdminHandler) Debug(c *gin.Context) {
	h.log.Warn("[VULNERABLE] Debug endpoint accessed")

	stack := string(debug.Stack())
	users, _, _ := h.userRepo.ListAll(1, 5)

	sampleUsers := make([]models.UserVulnerableResponse, len(users))
	for i, u := range users {
		sampleUsers[i] = u.ToVulnerableResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "[VULNERABLE] Debug endpoint — exposes internals",
		"security_mode": "vulnerable",
		"goroutine_stack": stack,        // VULN: internal goroutine stack
		"sample_users":  sampleUsers,   // VULN: user data with hashes
		"env": gin.H{                   // VULN: environment exposure
			"app_security_mode": security.GetMode(),
		},
		"owasp": []string{"API8: Security Misconfiguration", "API9: Improper Inventory Management"},
	})
}
