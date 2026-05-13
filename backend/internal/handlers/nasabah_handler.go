package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/repository"
	"pt-dana-sejahtera/internal/services"
)

type NasabahHandler struct {
	svc services.NasabahService
	log *zap.Logger
}

func NewNasabahHandler(svc services.NasabahService, log *zap.Logger) *NasabahHandler {
	return &NasabahHandler{svc: svc, log: log}
}

// Register godoc
// POST /api/v1/nasabah
// Requires role=nasabah; registers own profile.
func (h *NasabahHandler) Register(c *gin.Context) {
	var req models.CreateNasabahRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	result, err := h.svc.Create(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GetMyProfile godoc
// GET /api/v1/nasabah/me
func (h *NasabahHandler) GetMyProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	result, err := h.svc.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "nasabah profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch profile"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// GET /api/v1/nasabah/:id
// OWASP API1: ownership enforced in secure mode.
func (h *NasabahHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nasabah id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "nasabah not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusNotFound, gin.H{"error": "nasabah not found"}) // hide existence
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch nasabah"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

// List godoc
// GET /api/v1/nasabah?page=1&limit=10
// Admin/staff only in secure mode.
func (h *NasabahHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.svc.List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch nasabah list"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update godoc
// PUT /api/v1/nasabah/:id
func (h *NasabahHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nasabah id"})
		return
	}

	var req models.UpdateNasabahRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.Update(id, req, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "nasabah not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another nasabah's profile"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
// DELETE /api/v1/nasabah/:id  (admin only)
func (h *NasabahHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid nasabah id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	if err := h.svc.Delete(id, userID, role); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "nasabah not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		}
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
