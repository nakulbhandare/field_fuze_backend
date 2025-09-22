# Middleware

The Middleware layer provides cross-cutting concerns that apply to HTTP requests across the application. This includes authentication, authorization, CORS handling, request logging, and error recovery.

## 📋 Table of Contents

- [Overview](#overview)
- [Available Middleware](#available-middleware)
- [Authentication Middleware](#authentication-middleware)
- [CORS Middleware](#cors-middleware)
- [Logging Middleware](#logging-middleware)
- [Adding Custom Middleware](#adding-custom-middleware)
- [Middleware Chain](#middleware-chain)
- [Best Practices](#best-practices)

## Overview

Middleware in FieldFuze handles:
- **Authentication**: JWT token validation and user context
- **Authorization**: Permission-based access control with role validation
- **CORS**: Cross-Origin Resource Sharing configuration
- **Logging**: Structured request/response logging
- **Recovery**: Panic recovery and error handling
- **Security**: Request validation and security headers

All middleware follows the Gin framework's middleware pattern and can be applied globally or to specific routes.

## Available Middleware

### 1. Authentication Middleware (JWTManager)

**File**: `middelware/auth.go`

The most comprehensive middleware providing ultra-optimized authentication and authorization.

#### Features
- Ultra-fast permission checking with intelligent caching
- Zero-configuration auto-detection for APIs
- Advanced Go concurrency with goroutines and channels
- 8 core permissions system (read, write, delete, admin, manage, create, update, view)
- Thread-safe operations with sync.Map and atomic counters
- Real-time metrics and performance monitoring
- Context-aware authorization with timeout protection

#### Core Permissions
```go
type CorePermission string

const (
    PermissionRead   CorePermission = "read"     // GET operations and data retrieval
    PermissionWrite  CorePermission = "write"    // PATCH operations and data modification
    PermissionDelete CorePermission = "delete"   // DELETE operations and data removal
    PermissionAdmin  CorePermission = "admin"    // ALL access (highest level)
    PermissionManage CorePermission = "manage"   // create, update, delete operations
    PermissionCreate CorePermission = "create"   // POST operations and data creation
    PermissionUpdate CorePermission = "update"   // PUT operations and data updates
    PermissionView   CorePermission = "view"     // read-only access and reports
)
```

#### HTTP Method to Permission Mapping
```go
const (
    HTTPMethodGET    = "GET"    // → read/view permissions
    HTTPMethodPOST   = "POST"   // → create permissions
    HTTPMethodPUT    = "PUT"    // → update permissions
    HTTPMethodPATCH  = "PATCH"  // → write permissions
    HTTPMethodDELETE = "DELETE" // → delete permissions
)
```

#### Usage Examples

##### Standard Usage (Backward Compatible)
```go
// Apply authentication middleware
router.Use(jwtManager.AuthMiddleware())

// Require specific permission
router.GET("/users", jwtManager.RequirePermission("read"), handler)
router.POST("/users", jwtManager.RequirePermission("create"), handler)
router.DELETE("/users/:id", jwtManager.RequirePermission("delete"), handler)
```

##### Smart Auto-Detection (Zero Config)
```go
// Apply authentication middleware
router.Use(jwtManager.AuthMiddleware())

// Auto-detect permissions based on HTTP method
router.GET("/users", jwtManager.RequireSmartPermission(), handler)     // Auto: read
router.POST("/users", jwtManager.RequireSmartPermission(), handler)    // Auto: create
router.DELETE("/settings", jwtManager.RequireSmartPermission(), handler) // Auto: delete
```

##### Advanced Context-Aware Authorization
```go
// Require permission with additional context validation
router.PUT("/users/:id", jwtManager.RequireAdvancedPermission("update", map[string]string{
    "ownership": "required",
}), handler)

// Resource-specific permissions
router.GET("/users/:id", jwtManager.RequireResourcePermission("user_details"), handler)
```

##### Ownership-Based Access
```go
// Check if user owns the resource
router.GET("/users/:id", jwtManager.RequireOwnership(), handler)
router.PATCH("/users/:id", jwtManager.RequireOwnership(), handler)
```

#### Permission Cache Configuration
```go
const (
    DefaultCacheExpiry  = 5 * time.Minute  // Cache TTL
    DefaultCacheCleanup = 1 * time.Minute  // Cache cleanup interval
    
    PermissionEvalTimeout = 100 * time.Millisecond // Max evaluation time
    FastEvalTimeout       = 50 * time.Millisecond  // Fast path timeout
    
    MaxConcurrentEvals   = 10  // Concurrent permission evaluations
    DefaultChannelBuffer = 100 // Channel buffer size
)
```

#### JWT Claims Structure
The middleware extracts and validates JWT claims:

```go
type JWTClaims struct {
    UserID   string           `json:"user_id"`
    Email    string           `json:"email"`
    Username string           `json:"username"`
    Role     UserRole         `json:"role"`     // Backward compatibility
    Status   UserStatus       `json:"status"`
    Roles    []RoleAssignment `json:"roles"`    // Enhanced role system
    Context  UserContext      `json:"context"`  // User context
    
    jwt.RegisteredClaims
}
```

#### Middleware Chain Examples
```go
// In controller/controller.go:86-110

// Public routes (no auth)
user.POST("/register", c.User.Register)
user.POST("/login", c.User.Login)
user.POST("/validate", c.User.ValidateToken)

// Protected routes with auth + permissions
user.POST("/logout", 
    c.User.jwtManager.AuthMiddleware(), 
    c.User.Logout)

user.GET("/:id", 
    c.User.jwtManager.AuthMiddleware(), 
    c.User.jwtManager.RequireResourcePermission("user_details"), 
    c.User.GetUser)

user.GET("/list", 
    c.User.jwtManager.AuthMiddleware(), 
    c.User.jwtManager.RequireResourcePermission("user_list"), 
    c.User.GetUserList)

// Role assignment with admin permissions
user.POST("/:user_id/role/:role_id", 
    c.User.jwtManager.AuthMiddleware(), 
    c.User.jwtManager.RequireResourcePermission("role_assign"), 
    c.User.AssignRole)

// Infrastructure routes (admin only)
infra := v1.Group("/infrastructure", 
    c.User.jwtManager.AuthMiddleware(), 
    c.User.jwtManager.RequirePermission("admin"))
```

#### Performance Features
- **Intelligent Caching**: Permission results cached with TTL
- **Concurrent Evaluation**: Goroutines for parallel permission checks
- **Fast Path**: Optimized evaluation for common permissions
- **Memory Efficient**: sync.Map for thread-safe caching
- **Metrics**: Real-time performance monitoring

### 2. CORS Middleware

**File**: `middelware/cors.go`

Handles Cross-Origin Resource Sharing (CORS) configuration.

#### Structure
```go
type CORSMiddleware struct {
    config *models.Config
}

func NewCORSMiddleware(cfg *models.Config) *CORSMiddleware {
    return &CORSMiddleware{
        config: cfg,
    }
}
```

#### CORS Configuration
```go
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
```

#### Origin Validation
```go
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
```

#### Configuration Example
```json
{
    "cors_origins": [
        "https://app.fieldfuze.com",
        "https://admin.fieldfuze.com",
        "*.fieldfuze.com",
        "http://localhost:3000"
    ]
}
```

#### Usage
```go
// Apply CORS middleware globally
corsMiddleware := middelware.NewCORSMiddleware(config)
r.Use(corsMiddleware.CORS())
```

### 3. Logging Middleware

**File**: `middelware/logging.go`

Provides comprehensive request logging and error recovery.

#### Structure
```go
type LoggingMiddleware struct {
    logger logger.Logger
}

func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
    return &LoggingMiddleware{
        logger: log,
    }
}
```

#### Structured Request Logging
```go
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
```

#### Panic Recovery
```go
func (m *LoggingMiddleware) Recovery() gin.HandlerFunc {
    return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, recovered interface{}) {
        // Log the panic
        m.logger.Errorf("Panic recovered: %v", recovered)
        
        // Return 500 status
        c.JSON(500, gin.H{
            "error":   "Internal Server Error",
            "message": "An unexpected error occurred",
        })
        c.Abort()
    })
}
```

#### Standard Gin Logger
```go
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
```

#### Usage
```go
// Apply logging middleware
loggingMiddleware := middelware.NewLoggingMiddleware(logger.NewLogger(config.LogLevel, config.LogFormat))
r.Use(loggingMiddleware.StructuredLogger())
r.Use(loggingMiddleware.Recovery())
```

#### Log Output Examples

**Successful Request**:
```json
{
  "level": "info",
  "msg": "HTTP request completed successfully",
  "method": "GET",
  "path": "/api/v1/auth/user/123",
  "status": 200,
  "latency": "15.2ms",
  "ip": "192.168.1.100",
  "user_id": "user-123",
  "user_agent": "Mozilla/5.0...",
  "timestamp": "2024-01-15T10:30:45Z"
}
```

**Error Request**:
```json
{
  "level": "error",
  "msg": "HTTP request completed with error",
  "method": "POST",
  "path": "/api/v1/auth/user/role",
  "status": 500,
  "latency": "250.5ms",
  "ip": "192.168.1.100",
  "user_id": "user-123",
  "errors": "database connection failed",
  "timestamp": "2024-01-15T10:30:45Z"
}
```

## Adding Custom Middleware

### Step 1: Create Middleware File

```go
// middelware/rate_limit.go
package middelware

import (
    "net/http"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
)

type RateLimitMiddleware struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}

func NewRateLimitMiddleware(limit int, window time.Duration) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        
        m.mutex.Lock()
        defer m.mutex.Unlock()
        
        now := time.Now()
        windowStart := now.Add(-m.window)
        
        // Clean old requests
        if requests, exists := m.requests[clientIP]; exists {
            validRequests := []time.Time{}
            for _, reqTime := range requests {
                if reqTime.After(windowStart) {
                    validRequests = append(validRequests, reqTime)
                }
            }
            m.requests[clientIP] = validRequests
        }
        
        // Check rate limit
        if len(m.requests[clientIP]) >= m.limit {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": m.window.Seconds(),
            })
            c.Abort()
            return
        }
        
        // Add current request
        m.requests[clientIP] = append(m.requests[clientIP], now)
        
        c.Next()
    }
}
```

### Step 2: Register Middleware

```go
// In controller/controller.go
func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
    // Apply global middleware
    corsMiddleware := middelware.NewCORSMiddleware(config)
    r.Use(corsMiddleware.CORS())
    
    loggingMiddleware := middelware.NewLoggingMiddleware(logger.NewLogger(config.LogLevel, config.LogFormat))
    r.Use(loggingMiddleware.StructuredLogger())
    r.Use(loggingMiddleware.Recovery())
    
    // Apply rate limiting
    rateLimitMiddleware := middelware.NewRateLimitMiddleware(100, time.Minute)
    r.Use(rateLimitMiddleware.RateLimit())
    
    // ... rest of route setup
}
```

### Step 3: Route-Specific Middleware

```go
// Apply middleware to specific routes
api := r.Group("/api", rateLimitMiddleware.RateLimit())
{
    api.POST("/upload", uploadHandler)
    api.POST("/process", processHandler)
}

// Apply multiple middleware to a single route
r.GET("/admin/users", 
    jwtManager.AuthMiddleware(),
    jwtManager.RequirePermission("admin"),
    rateLimitMiddleware.RateLimit(),
    adminUserHandler)
```

## Middleware Chain

### Global Middleware Order
The order of middleware application is important:

```go
func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
    // 1. CORS (must be first for preflight requests)
    corsMiddleware := middelware.NewCORSMiddleware(config)
    r.Use(corsMiddleware.CORS())
    
    // 2. Logging (track all requests)
    loggingMiddleware := middelware.NewLoggingMiddleware(logger.NewLogger(config.LogLevel, config.LogFormat))
    r.Use(loggingMiddleware.StructuredLogger())
    r.Use(loggingMiddleware.Recovery())
    
    // 3. Rate limiting (before auth to prevent abuse)
    if config.RateLimitRequestsPerMinute > 0 {
        rateLimitMiddleware := middelware.NewRateLimitMiddleware(config.RateLimitRequestsPerMinute, time.Minute)
        r.Use(rateLimitMiddleware.RateLimit())
    }
    
    // 4. Security headers
    r.Use(securityHeadersMiddleware())
    
    // 5. Route-specific middleware applied per group/route
    v1 := r.Group(basePath)
    // Auth middleware applied per route as needed
}
```

### Route-Specific Chain
```go
// Example: Protected user management route
user.PATCH("/update/:id", 
    c.User.jwtManager.AuthMiddleware(),                              // 1. Authentication
    c.User.jwtManager.RequireResourcePermission("user_update"),     // 2. Authorization
    validationMiddleware.ValidateUserUpdate(),                      // 3. Input validation
    auditMiddleware.LogUserChanges(),                              // 4. Audit logging
    c.User.UpdateUser)                                             // 5. Handler
```

## Best Practices

### 1. Middleware Order
- **CORS first**: Handle preflight OPTIONS requests
- **Recovery early**: Catch panics before they propagate
- **Logging early**: Track all requests including failed ones
- **Authentication before authorization**: Verify identity before permissions
- **Validation after auth**: Only validate for authenticated users

### 2. Performance Considerations
- **Cache permissions**: Use intelligent caching for expensive permission checks
- **Async logging**: Don't block requests for logging operations
- **Rate limiting**: Protect against abuse and DoS attacks
- **Connection pooling**: Reuse database connections

### 3. Security Best Practices
- **Validate all inputs**: Never trust user input
- **Use HTTPS**: Enforce secure connections
- **JWT security**: Proper token validation and expiration
- **Error handling**: Don't leak sensitive information in errors
- **Audit logging**: Track security-relevant operations

### 4. Error Handling
```go
func secureErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Log full error for debugging
                logger.Errorf("Panic in request: %v", err)
                
                // Return generic error to client
                c.JSON(http.StatusInternalServerError, models.APIResponse{
                    Status:  "error",
                    Code:    http.StatusInternalServerError,
                    Message: "Internal server error",
                    Error: &models.APIError{
                        Type: "ServerError",
                        Details: "An unexpected error occurred",
                    },
                })
                c.Abort()
            }
        }()
        
        c.Next()
    }
}
```

### 5. Context Management
```go
func contextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Add request ID for tracing
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        // Add timeout context
        ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
        defer cancel()
        
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

### 6. Conditional Middleware
```go
func conditionalAuth(requireAuth bool) gin.HandlerFunc {
    return func(c *gin.Context) {
        if requireAuth {
            // Apply authentication
            jwtManager.AuthMiddleware()(c)
            if c.IsAborted() {
                return
            }
        }
        c.Next()
    }
}
```

### 7. Middleware Testing
```go
func TestAuthMiddleware(t *testing.T) {
    // Setup
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // Mock request with valid JWT
    c.Request = httptest.NewRequest("GET", "/test", nil)
    c.Request.Header.Set("Authorization", "Bearer valid_jwt_token")
    
    // Test middleware
    jwtManager.AuthMiddleware()(c)
    
    // Assertions
    assert.False(t, c.IsAborted())
    assert.NotNil(t, c.Get("jwt_claims"))
}
```

---

**Related Documentation**: [Controllers](CONTROLLERS.md) | [API Reference](API.md) | [Development Guide](DEVELOPMENT.md) | [Security](SECURITY.md)