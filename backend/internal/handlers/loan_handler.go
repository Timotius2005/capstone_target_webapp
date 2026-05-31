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

type LoanHandler struct {
	svc services.LoanService
	log *zap.Logger
}

func NewLoanHandler(svc services.LoanService, log *zap.Logger) *LoanHandler {
	return &LoanHandler{svc: svc, log: log}
}

// Apply godoc
// POST /api/v1/loans
// OWASP API4 + API6: throttled in secure; unlimited in vulnerable.
func (h *LoanHandler) Apply(c *gin.Context) {
	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		return
	}

	result, err := h.svc.Apply(req, userID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrForbidden) {
			status = http.StatusForbidden
		} else if errors.Is(err, repository.ErrNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// List godoc
// GET /api/v1/loans?page=1&limit=10
// OWASP API1 + API4: ownership-filtered (secure); all records (vulnerable).
func (h *LoanHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.List(userID, role, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch loans"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// GET /api/v1/loans/:id
// OWASP API1: ownership check (secure); no check (vulnerable).
func (h *LoanHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.GetByID(id, userID, role)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrForbidden) {
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch loan"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UpdateStatus godoc
// PATCH /api/v1/loans/:id/status
// OWASP API3 (BOPLA): staff cannot approve (secure); free update (vulnerable).
func (h *LoanHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	var req models.UpdateLoanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.UpdateStatus(id, req, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

// Approve godoc
// POST /api/v1/loans/:id/approve
// OWASP API5: admin only (secure); anyone (vulnerable).
func (h *LoanHandler) Approve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.Approve(id, userID, role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required to approve loans"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

// Reject godoc
// POST /api/v1/loans/:id/reject
func (h *LoanHandler) Reject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid loan id"})
		return
	}

	var body struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&body)

	userID, _ := uuid.Parse(c.GetString("user_id"))
	role := c.GetString("role")

	result, err := h.svc.Reject(id, userID, role, body.Notes)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "loan not found"})
		case errors.Is(err, repository.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, result)
}

// ApplyPublic godoc — POST /api/v0/loans
// TODO: Vulnerability Injection Point — OWASP API6 (Unrestricted Business Flow)
// No authentication required, no pending-loan cap — unlimited loan creation.
func (h *LoanHandler) ApplyPublic(c *gin.Context) {
	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Use a fresh anonymous user ID so the service auto-creates a nasabah profile.
	anonID := uuid.New()
	h.log.Warn("[VULNERABLE] v0 public loan application — no auth, no cap (A06)",
		zap.String("anon_id", anonID.String()),
		zap.Float64("amount", req.Amount),
	)
	result, err := h.svc.Apply(req, anonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// ListPublic godoc — OWASP API9: deprecated v0 route, no auth required.
// TODO: Vulnerability Injection Point — OWASP API9 (Improper Inventory Management)
func (h *LoanHandler) ListPublic(c *gin.Context) {
	h.log.Warn("[VULNERABLE] v0 public loan list accessed — no authentication required")
	result, err := h.svc.List(uuid.Nil, models.RoleAdmin, 1, 9999)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}
