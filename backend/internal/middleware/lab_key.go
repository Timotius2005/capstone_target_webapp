package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// LabKeyRequired provides optional token-based access control for public lab endpoints.
// Protection is ENABLED only when the LAB_KEY environment variable is set.
// When active, every request must carry: X-LAB-KEY: <value>
//
// To enable:  set LAB_KEY=pentest123 in .env or docker-compose environment.
// To disable: leave LAB_KEY unset (all requests pass through).
func LabKeyRequired() gin.HandlerFunc {
	labKey := os.Getenv("LAB_KEY")
	return func(c *gin.Context) {
		if labKey == "" {
			// Protection disabled — allow all requests unconditionally.
			c.Next()
			return
		}
		if c.GetHeader("X-LAB-KEY") != labKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "X-LAB-KEY header missing or invalid",
				"hint":  "include the header: X-LAB-KEY: <lab_key>",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
