package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"pt-dana-sejahtera/internal/usecase"
)

type NasabahHandler struct {
	nasabahUseCase *usecase.NasabahUseCase
}

func NewNasabahHandler(nasabahUseCase *usecase.NasabahUseCase) *NasabahHandler {
	return &NasabahHandler{nasabahUseCase: nasabahUseCase}
}

func (h *NasabahHandler) GetAllNasabah(c *gin.Context) {
	nasabah, err := h.nasabahUseCase.GetAllNasabah()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: SECURITY VULNERABILITY - Mass Assignment
	// Should not return all fields, especially sensitive data
	c.JSON(http.StatusOK, nasabah)
}

func (h *NasabahHandler) GetNasabahByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	nasabah, err := h.nasabahUseCase.GetNasabahByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nasabah not found"})
		return
	}

	c.JSON(http.StatusOK, nasabah)
}

func (h *NasabahHandler) CreateNasabah(c *gin.Context) {
	var req usecase.CreateNasabahRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: SECURITY VULNERABILITY - No Input Validation
	// Should validate NIK format, phone number, etc.
	err := h.nasabahUseCase.CreateNasabah(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Nasabah created successfully"})
}