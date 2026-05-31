package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/security"
)

// bucket holds a simple token-bucket counter per IP.
type bucket struct {
	mu        sync.Mutex
	count     int
	resetAt   time.Time
	windowSec int
	maxReq    int
}

func (b *bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	if now.After(b.resetAt) {
		b.count = 0
		b.resetAt = now.Add(time.Duration(b.windowSec) * time.Second)
	}
	if b.count >= b.maxReq {
		return false
	}
	b.count++
	return true
}

// ipStore is the global in-memory rate limit store.
var ipStore = &sync.Map{}

// RateLimit returns a middleware that limits requests per IP.
// OWASP #4 Secure: unrestricted resource consumption protection.
// OWASP A09 Vulnerable: no rate limiting — brute-force goes undetected.
func RateLimit(maxReq, windowSec int, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.IsVulnerable() {
			// TODO: Vulnerability Injection Point — OWASP API4 / A09 (Security Logging Failures)
			// Vulnerable mode: global rate limiting disabled — brute-force undetected.
			c.Next()
			return
		}

		ip := c.ClientIP()
		// Use the actual URL path (not route pattern) so each distinct URL
		// has its own counter — BOLA tests on different UUIDs are independent.
		key := ip + c.Request.URL.Path

		val, _ := ipStore.LoadOrStore(key, &bucket{
			resetAt:   time.Now().Add(time.Duration(windowSec) * time.Second),
			windowSec: windowSec,
			maxReq:    maxReq,
		})
		b := val.(*bucket)

		if !b.Allow() {
			log.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.FullPath()),
			)
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded — please slow down",
			})
			return
		}
		c.Next()
	}
}

// LoginRateLimit applies stricter limits on login endpoint.
// OWASP #2 Secure: prevent brute-force.
func LoginRateLimit(log *zap.Logger) gin.HandlerFunc {
	return RateLimit(5, 60, log) // 5 requests per 60 seconds per IP
}

// GlobalRateLimit enforces a per-IP per-URL rate limit in secure mode.
// 30 req/min per URL is permissive enough for normal API usage but blocks
// scanners that flood a single endpoint with 50+ rapid requests (A04).
func GlobalRateLimit(log *zap.Logger) gin.HandlerFunc {
	return RateLimit(30, 60, log)
}
