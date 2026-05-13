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

type TransactionHandler struct {
	svc services.TransactionService
	log *zap.Logger
}

func NewTransactionHandler(svc services.TransactionService, log *zap.Logger) *TransactionHandler {
	return &TransactionHandler{svc: svc, log: log}
}

// Create godoc
// POST /api/v1/transactions
// OWASP API5: admin/staff only (secure); anyone (vulnerable).
func (h *TransactionHandler) Create(c *gin.Context) {
	var req models.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.Create(req, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, result)
}

// ListByLoan godoc
// GET /api/v1/loans/:id/transactions
// OWASP API1: ownership enforced (secure).
func (h *TransactionHandler) ListByLoan(c *gin.Context) {
	loanID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.ListByLoan(loanID, userID, role, page, limit)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// List godoc
// GET /api/v1/transactions (admin/staff only)
func (h *TransactionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.svc.List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, result)
}
