package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pt-dana-sejahtera/internal/models"
	"pt-dana-sejahtera/internal/security"
	"pt-dana-sejahtera/internal/services"
)

const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
	ContextRole     = "role"
)

// AuthRequired validates the Bearer JWT and injects claims into context.
func AuthRequired(authSvc services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header missing or malformed",
			})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := authSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// AdminOnly enforces role=admin in secure mode.
// OWASP API5 Secure: role-based function level authorization.
// OWASP API5 Vulnerable [A07]: no role check — any authenticated user reaches admin routes.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.IsVulnerableFor(security.CategoryA07) {
			// TODO: Vulnerability Injection Point — OWASP API5 / A07 (Broken Function Level Authorization)
			// A07 enabled: any authenticated user can reach admin endpoints.
			c.Next()
			return
		}
		role, _ := c.Get(ContextRole)
		if role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}
		c.Next()
	}
}

// RoleCheck allows the specified roles only.
// Always enforced regardless of mode (not an injection point).
func RoleCheck(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists || !allowed[role.(string)] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}
		c.Next()
	}
}
