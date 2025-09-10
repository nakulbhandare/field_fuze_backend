package middelware

import (
	"time"

	"github.com/gin-gonic/gin"
	"fieldfuze-backend/utils/logger"
)

// LoggingMiddleware provides request logging
type LoggingMiddleware struct {
	logger logger.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: log,
	}
}

// RequestLogger returns a gin.HandlerFunc for logging HTTP requests
func (m *LoggingMiddleware) RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Custom log format
			m.logger.Infof("[%s] %s %s %d %s %s %s",
				param.TimeStamp.Format("2006-01-02 15:04:05"),
				param.ClientIP,
				param.Method,
				param.StatusCode,
				param.Latency,
				param.Path,
				param.ErrorMessage,
			)
			return ""
		},
		SkipPaths: []string{"/health", "/metrics"}, // Skip health checks
	})
}

// StructuredLogger provides structured logging for requests
func (m *LoggingMiddleware) StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID from context if available
		userID, _ := c.Get("user_id")

		// Log request details
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"status":     c.Writer.Status(),
			"latency":    latency,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		if userID != nil {
			fields["user_id"] = userID
		}

		// Add error details if any
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			m.logger.Errorf("HTTP request completed with error: %+v", fields)
		} else if c.Writer.Status() >= 400 {
			m.logger.Warnf("HTTP request completed with client error: %+v", fields)
		} else {
			m.logger.Infof("HTTP request completed successfully: %+v", fields)
		}
	}
}

// Recovery middleware with logging
func (m *LoggingMiddleware) Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, recovered interface{}) {
		// Log the panic
		m.logger.Errorf("Panic recovered: %v", recovered)

		// Return 500 status
		c.JSON(500, gin.H{
			"error":   "Internal Server Error",
			"message": "An unexpected error occurred",
		})
	})
}
