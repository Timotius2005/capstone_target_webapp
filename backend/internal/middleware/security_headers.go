package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/security"
)

// SecureHeaders adds OWASP-recommended security headers.
// OWASP #8 Secure: security misconfiguration protection.
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.IsSecure() {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
			c.Header("Cache-Control", "no-store")
			// Remove fingerprinting headers
			c.Header("Server", "")
		} else {
			// TODO: Vulnerability Injection Point — OWASP API #8 (Security Misconfiguration)
			// Vulnerable mode: exposes server version, no security headers
			c.Header("Server", "Go/gin-v1.9 pt-dana-sejahtera/1.0")
			c.Header("X-Powered-By", "Go 1.21")
			c.Header("X-Debug-Mode", "true")
			c.Header("X-Security-Mode", "vulnerable")
		}
		c.Next()
	}
}

// ErrorHandler provides mode-aware error responses.
// OWASP #8 Secure: no stack traces in production.
func ErrorHandler(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Sprintf("%v", r)
				stack := debug.Stack()

				log.Error("Panic recovered",
					zap.String("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				if security.IsVulnerable() {
					// TODO: Vulnerability Injection Point — OWASP API #8 (Security Misconfiguration)
					// Exposes full stack trace and internal error details
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":       "internal server error",
						"detail":      err,
						"stack_trace": string(stack), // VULN: leaks internal paths & logic
						"debug":       true,
					})
				} else {
					// Secure: generic message, no internal details
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "an internal error occurred",
					})
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
