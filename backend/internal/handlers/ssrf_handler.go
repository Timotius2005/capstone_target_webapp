package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/services"
)

type SSRFHandler struct {
	svc services.ExternalService
	log *zap.Logger
}

func NewSSRFHandler(svc services.ExternalService, log *zap.Logger) *SSRFHandler {
	return &SSRFHandler{svc: svc, log: log}
}

// Fetch godoc
// POST /api/v1/internal/fetch
//
// OWASP API7: SSRF vulnerability demonstration.
// Secure:   domain whitelist + private IP block.
// Vulnerable: fetch any URL including internal services.
//
// Test vulnerable: {"url": "http://169.254.169.254/latest/meta-data/"}
// Test secure:     {"url": "https://jsonplaceholder.typicode.com/todos/1"}
func (h *SSRFHandler) Fetch(c *gin.Context) {
	var req services.FetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Fetch(req)
	if err != nil {
		h.log.Warn("External fetch failed",
			zap.String("url", req.URL),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
