package middelware

import (
	"fieldfuze-backend/models"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides CORS handling
type CORSMiddleware struct {
	config *models.Config
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(cfg *models.Config) *CORSMiddleware {
	return &CORSMiddleware{
		config: cfg,
	}
}

// CORS returns a gin.HandlerFunc for handling CORS
func (m *CORSMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if m.isOriginAllowed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	// If no origin specified, allow it
	if origin == "" {
		return true
	}

	// Check against configured origins
	for _, allowedOrigin := range m.config.CORSOrigins {
		// Allow all origins if * is configured
		if allowedOrigin == "*" {
			return true
		}

		// Exact match
		if allowedOrigin == origin {
			return true
		}

		// Wildcard subdomain matching (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:] // Remove "*."
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}
