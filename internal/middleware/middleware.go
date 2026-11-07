package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs method, path, status, and elapsed time after each request.
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http request", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration", time.Since(started).String())
	}
}

// ErrorLogger writes Gin errors and failed HTTP statuses without exposing internals.
func ErrorLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			logger.Error("request error", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "error", c.Errors.String())
			return
		}
		if c.Writer.Status() >= 500 {
			logger.Error("http server error", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status())
		}
	}
}

// SecurityHeaders adds baseline browser security headers to every response.
func SecurityHeaders(development bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "SAMEORIGIN")
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		header.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if !development {
			header.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'self'")
		}
		c.Next()
	}
}
