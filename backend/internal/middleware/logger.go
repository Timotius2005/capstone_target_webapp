package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pt-dana-sejahtera/internal/security"
)

// RequestLogger logs each HTTP request with structured fields.
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("security_mode", security.GetMode()),
		}

		if query != "" {
			fields = append(fields, zap.String("query", query))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors.Errors() {
				fields = append(fields, zap.String("error", e))
			}
			log.Error("Request completed with errors", fields...)
			return
		}

		switch {
		case status >= 500:
			log.Error("Request", fields...)
		case status >= 400:
			log.Warn("Request", fields...)
		default:
			log.Info("Request", fields...)
		}
	}
}
