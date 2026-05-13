package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"pt-dana-sejahtera/internal/security"
)

const (
	// MaxSecureBodySize: 1 MB in secure mode.
	MaxSecureBodySize int64 = 1 << 20
	// MaxVulnerableBodySize: 500 MB — effectively no limit.
	// TODO: Vulnerability Injection Point — OWASP API #4 (Unrestricted Resource Consumption)
	MaxVulnerableBodySize int64 = 500 << 20
)

// RequestSizeLimit enforces a maximum request body size.
// OWASP #4 Secure: unrestricted resource consumption protection.
func RequestSizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.IsVulnerable() {
			// TODO: Vulnerability Injection Point — OWASP API #4
			// Vulnerable: allows enormous payloads — DoS via memory exhaustion
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxVulnerableBodySize)
		} else {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxSecureBodySize)
		}

		c.Next()

		if c.Request.Body != nil {
			if err := c.Request.Body.Close(); err == nil {
				return
			}
		}
	}
}
