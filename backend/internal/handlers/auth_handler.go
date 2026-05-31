package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/services"
)

type AuthHandler struct {
	authSvc services.AuthService
	log     *zap.Logger
}

func NewAuthHandler(authSvc services.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, log: log}
}

// Register godoc
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authSvc.Register(req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registration successful",
		"token":   result.Token,
		"user":    result.User,
	})
}

// Login godoc
// POST /api/v1/auth/login
// OWASP API2: behavior (bcrypt/rate-limit vs plain-text) set by APP_SECURITY_MODE.
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.authSvc.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefreshToken godoc
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newToken, err := h.authSvc.RefreshToken(body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// Me godoc
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"user_id":  c.GetString("user_id"),
		"username": c.GetString("username"),
		"role":     c.GetString("role"),
	})
}
