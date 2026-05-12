package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/usecase"
)

type LoanHandler struct {
	loanUseCase *usecase.LoanUseCase
}

func NewLoanHandler(loanUseCase *usecase.LoanUseCase) *LoanHandler {
	return &LoanHandler{loanUseCase: loanUseCase}
}

func (h *LoanHandler) GetAllLoans(c *gin.Context) {
	loans, err := h.loanUseCase.GetAllLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loans)
}

func (h *LoanHandler) GetLoanByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	loan, err := h.loanUseCase.GetLoanByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	c.JSON(http.StatusOK, loan)
}

func (h *LoanHandler) CreateLoan(c *gin.Context) {
	var req usecase.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: SECURITY VULNERABILITY - No Rate Limiting
	// Should implement rate limiting for loan creation
	err := h.loanUseCase.CreateLoan(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Loan created successfully"})
}