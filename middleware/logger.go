package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			b := make([]byte, 8)
			if _, err := rand.Read(b); err == nil {
				requestID = hex.EncodeToString(b)
			}
		}
		c.Header("X-Request-Id", requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()
		size := max(c.Writer.Size(), 0)

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.Int("size", size),
		}

		switch {
		case status >= 500:
			logger.Error("request", attrs...)
		case status >= 400:
			logger.Warn("request", attrs...)
		default:
			logger.Info("request", attrs...)
		}
	}
}
