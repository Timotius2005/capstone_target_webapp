package middleware

import (
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"pt-dana-sejahtera/internal/security"
)

// AllowedOrigins defines the CORS whitelist for secure mode.
// Extra origins (e.g. the Vercel frontend) can be added at runtime via the
// FRONTEND_ORIGIN env var — comma-separated for multiple values — so the
// hybrid-cloud deployment works without recompiling.
var AllowedOrigins = func() []string {
	base := []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"https://dana-sejahtera.example.com",
	}
	if extra := os.Getenv("FRONTEND_ORIGIN"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				base = append(base, o)
			}
		}
	}
	return base
}()

// CORS handles cross-origin resource sharing.
// OWASP #8 Secure: properly configured CORS.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if security.IsVulnerable() {
			// TODO: Vulnerability Injection Point — OWASP API #8 (Security Misconfiguration)
			// Wildcard CORS — allows any origin to make credentialed requests
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Expose-Headers", "*")
		} else {
			// Secure: restrict to known origins
			if origin != "" && slices.Contains(AllowedOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, ngrok-skip-browser-warning, X-LAB-KEY")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
