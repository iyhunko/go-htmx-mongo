package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger is a middleware that logs HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		slog.Info("Request processed",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
		)
	}
}
