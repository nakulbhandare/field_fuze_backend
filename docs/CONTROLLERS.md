# Controllers

The Controllers layer handles HTTP requests and responses, orchestrating between the middleware, services, and the client. This layer is responsible for request validation, response formatting, and delegating business logic to appropriate services.

## 📋 Table of Contents

- [Overview](#overview)
- [Controller Structure](#controller-structure)
- [Available Controllers](#available-controllers)
- [Adding New Controllers](#adding-new-controllers)
- [Request/Response Patterns](#requestresponse-patterns)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

Controllers in FieldFuze follow a clean architecture pattern where:
- Controllers handle HTTP concerns (request parsing, response formatting)
- Business logic is delegated to Services
- Database operations are handled by Repositories
- Cross-cutting concerns are managed by Middleware

## Controller Structure

### Main Controller Factory
**File**: `controller/controller.go:25-43`

```go
type Controller struct {
    User           *UserController
    Role           *RoleController
    Infrastructure *InfrastructureController
}

func NewController(ctx context.Context, cfg *models.Config, log logger.Logger) *Controller {
    // Initialize dependencies
    dbclient, err := dal.NewDynamoDBClient(cfg, log)
    userRepo := repository.NewUserRepository(dbclient, cfg, log)
    roleRepo := repository.NewRoleRepository(dbclient, cfg, log)
    jwtManager := middelware.NewJWTManager(cfg, log, userRepo)
    
    // Initialize services
    roleService := services.NewRoleService(roleRepo, log)
    infraService := services.NewInfrastructureService(ctx, dbclient, log, cfg)
    
    return &Controller{
        User:           NewUserController(ctx, userRepo, log, jwtManager),
        Role:           NewRoleController(ctx, roleService, log),
        Infrastructure: NewInfrastructureController(ctx, infraService, log),
    }
}
```

### Route Registration
**File**: `controller/controller.go:45-137`

```go
func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
    // Apply global middleware
    corsMiddleware := middelware.NewCORSMiddleware(config)
    r.Use(corsMiddleware.CORS())
    
    loggingMiddleware := middelware.NewLoggingMiddleware(logger.NewLogger(config.LogLevel, config.LogFormat))
    r.Use(loggingMiddleware.StructuredLogger())
    r.Use(loggingMiddleware.Recovery())
    
    v1 := r.Group(basePath)
    
    // Health check endpoint
    v1.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":  "healthy",
            "version": "1.0.0",
            "service": "FieldFuze Backend",
        })
    })
    
    // Setup route groups
    c.setupUserRoutes(v1)
    c.setupRoleRoutes(v1)
    c.setupInfrastructureRoutes(v1)
    
    // Start server
    srv := &http.Server{
        Addr:    config.AppHost + ":" + config.AppPort,
        Handler: r,
    }
    return srv.ListenAndServe()
}
```

## Available Controllers

### 1. UserController

**File**: `controller/user_controller.go`

Handles user authentication, registration, and management operations.

#### Structure
```go
type UserController struct {
    ctx        context.Context
    userRepo   *repository.UserRepository
    jwtManager *middelware.JWTManager
    logger     logger.Logger
}
```

#### Key Methods

##### User Registration
```go
// POST /api/v1/auth/user/register
func (h *UserController) Register(c *gin.Context) {
    var req models.User
    if err := c.ShouldBindJSON(&req); err != nil {
        // Handle validation error
        return
    }
    
    user, err := h.userRepo.CreateUser(h.ctx, &req)
    if err != nil {
        // Handle creation error
        return
    }
    
    c.JSON(http.StatusCreated, models.APIResponse{
        Status:  "success",
        Code:    http.StatusCreated,
        Message: "User registered successfully",
        Data:    user,
    })
}
```

##### User Authentication
```go
// POST /api/v1/auth/user/login
func (h *UserController) Login(c *gin.Context) {
    // Delegate to JWT manager for authentication
    h.jwtManager.HandleLogin(c)
}
```

##### Get User Details
```go
// GET /api/v1/auth/user/:id
func (h *UserController) GetUser(c *gin.Context) {
    userID := c.Param("id")
    users, err := h.userRepo.GetUser(userID)
    
    if err != nil {
        // Handle error
        return
    }
    
    if len(users) == 0 {
        c.JSON(http.StatusNotFound, models.APIResponse{
            Status:  "error",
            Code:    http.StatusNotFound,
            Message: "User not found",
        })
        return
    }
    
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "User details retrieved successfully",
        Data:    users[0],
    })
}
```

##### Role Assignment
```go
// POST /api/v1/auth/user/{user_id}/role/{role_id}
func (h *UserController) AssignRole(c *gin.Context) {
    userID := c.Param("user_id")
    roleID := c.Param("role_id")
    
    // Validate parameters
    if userID == "" || roleID == "" {
        // Return validation error
        return
    }
    
    // Assign role
    updatedUser, err := h.userRepo.AssignRoleToUser(h.ctx, userID, roleID)
    if err != nil {
        // Handle specific errors (conflict, not found, etc.)
        return
    }
    
    // Clear permission cache after role change
    h.invalidateUserPermissions(userID, "role_assignment")
    
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Role assigned successfully",
        Data:    updatedUser,
    })
}
```

#### Routes Overview
```go
// Public routes (no auth required)
user.POST("/register", c.User.Register)
user.POST("/login", c.User.Login)
user.POST("/token", c.User.GenerateToken)
user.POST("/validate", c.User.ValidateToken)

// Protected routes (auth + permissions required)
user.POST("/logout", c.User.jwtManager.AuthMiddleware(), c.User.Logout)
user.GET("/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_details"), c.User.GetUser)
user.GET("/list", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_list"), c.User.GetUserList)
user.PATCH("/update/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_update"), c.User.UpdateUser)

// Role assignment routes
user.POST("/:user_id/role/:role_id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_assign"), c.User.AssignRole)
user.DELETE("/:user_id/role/:role_id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_assign"), c.User.DetachRole)
```

### 2. RoleController

**File**: `controller/role_controller.go`

Manages role creation, updates, and assignment operations.

#### Structure
```go
type RoleController struct {
    ctx         context.Context
    roleService *services.RoleService
    logger      logger.Logger
    validator   *validator.Validate
}
```

#### Key Methods

##### Create Role
```go
// POST /api/v1/auth/user/role
func (h *RoleController) CreateRole(c *gin.Context) {
    var req models.RoleAssignment
    if err := c.ShouldBindJSON(&req); err != nil {
        // Handle binding error
        return
    }
    
    // Validate using struct validation
    if err := h.validator.Struct(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Status:  "error",
            Code:    http.StatusBadRequest,
            Message: "Validation failed",
            Error: &models.APIError{
                Type:    "ValidationError",
                Details: h.formatValidationErrors(err),
            },
        })
        return
    }
    
    // Extract JWT claims for audit
    claims, exists := c.Get("jwt_claims")
    if !exists {
        // Handle auth error
        return
    }
    
    jwtClaims := claims.(*models.JWTClaims)
    role, err := h.roleService.CreateRole(h.ctx, &req, jwtClaims.UserID)
    
    c.JSON(http.StatusCreated, models.APIResponse{
        Status:  "success",
        Code:    http.StatusCreated,
        Message: "Role created successfully",
        Data:    role,
    })
}
```

##### Get Roles with Pagination
```go
// GET /api/v1/auth/user/role
func (h *RoleController) GetRoles(c *gin.Context) {
    // Parse query parameters
    status := c.Query("status")
    page := 1
    limit := 10
    
    if pageParam := c.Query("page"); pageParam != "" {
        if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
            page = p
        }
    }
    
    if limitParam := c.Query("limit"); limitParam != "" {
        if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }
    
    // Get roles based on status filter
    var roles []*models.RoleAssignment
    var err error
    
    if status != "" {
        roles, err = h.roleService.GetRoleAssignmentsByStatus(status)
    } else {
        roles, err = h.roleService.GetRoleAssignments()
    }
    
    // Apply pagination
    total := len(roles)
    totalPages := (total + limit - 1) / limit
    offset := (page - 1) * limit
    
    var paginatedRoles []*models.RoleAssignment
    if offset < total {
        end := offset + limit
        if end > total {
            end = total
        }
        paginatedRoles = roles[offset:end]
    }
    
    responseData := map[string]interface{}{
        "roles": paginatedRoles,
        "pagination": map[string]interface{}{
            "page":         page,
            "limit":        limit,
            "total":        total,
            "total_pages":  totalPages,
            "has_next":     page < totalPages,
            "has_previous": page > 1,
        },
    }
    
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Roles retrieved successfully",
        Data:    responseData,
    })
}
```

#### Validation Helper
```go
func (h *RoleController) formatValidationErrors(err error) string {
    var errorMessages []string
    
    if validationErrors, ok := err.(validator.ValidationErrors); ok {
        for _, fieldError := range validationErrors {
            switch fieldError.Tag() {
            case "required":
                errorMessages = append(errorMessages, fieldError.Field()+" is required")
            case "min":
                errorMessages = append(errorMessages, fieldError.Field()+" must be at least "+fieldError.Param()+" characters/items")
            case "max":
                errorMessages = append(errorMessages, fieldError.Field()+" must be at most "+fieldError.Param()+" characters/items")
            default:
                errorMessages = append(errorMessages, fieldError.Field()+" is invalid")
            }
        }
    }
    
    return strings.Join(errorMessages, "; ")
}
```

### 3. InfrastructureController

**File**: `controller/infrastructure_controller.go`

Manages infrastructure worker operations and health monitoring.

#### Structure
```go
type InfrastructureController struct {
    ctx     context.Context
    service *services.InfrastructureService
    logger  logger.Logger
}
```

#### Key Methods

##### Worker Status
```go
// GET /api/v1/infrastructure/worker/status
func (h *InfrastructureController) GetWorkerStatus(c *gin.Context) {
    workerStatus, err := h.service.GetWorkerStatus(h.ctx)
    if err != nil {
        h.logger.Error("Failed to get worker status", err)
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Status:  "error",
            Code:    http.StatusInternalServerError,
            Message: "Failed to retrieve worker status",
            Error: &models.APIError{
                Type:    "WorkerError",
                Details: err.Error(),
            },
        })
        return
    }
    
    // Map worker status to HTTP status
    httpStatus, apiStatus := h.mapWorkerStatusToHTTP(workerStatus)
    message := h.getStatusMessage(workerStatus)
    
    c.JSON(httpStatus, models.APIResponse{
        Status:  apiStatus,
        Code:    httpStatus,
        Message: message,
        Data:    workerStatus,
    })
}
```

##### Worker Restart
```go
// POST /api/v1/infrastructure/worker/restart
func (h *InfrastructureController) RestartWorker(c *gin.Context) {
    var restartRequest models.WorkerRestartRequest
    if err := c.ShouldBindJSON(&restartRequest); err != nil {
        // Use defaults if no body provided
        restartRequest = models.WorkerRestartRequest{Force: false}
    }
    
    result, err := h.service.RestartWorker(h.ctx, restartRequest.Force)
    if err != nil {
        // Handle conflict (worker running)
        if strings.Contains(err.Error(), "worker is running") {
            c.JSON(http.StatusConflict, models.APIResponse{
                Status:  "error",
                Code:    http.StatusConflict,
                Message: "Worker is currently running",
                Error: &models.APIError{
                    Type:    "ConflictError",
                    Details: "Worker is currently running. Use force=true to restart anyway",
                },
            })
            return
        }
        
        // Handle other errors
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Status:  "error",
            Code:    http.StatusInternalServerError,
            Message: "Failed to restart worker",
            Error: &models.APIError{
                Type:    "WorkerError",
                Details: err.Error(),
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Worker restart initiated successfully",
        Data:    result,
    })
}
```

## Adding New Controllers

### Step 1: Create Controller File

Create a new file in the `controller/` directory:

```go
// controller/your_controller.go
package controller

import (
    "context"
    "fieldfuze-backend/services"
    "fieldfuze-backend/utils/logger"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type YourController struct {
    ctx     context.Context
    service *services.YourService
    logger  logger.Logger
}

func NewYourController(ctx context.Context, service *services.YourService, logger logger.Logger) *YourController {
    return &YourController{
        ctx:     ctx,
        service: service,
        logger:  logger,
    }
}

// Your endpoint handlers
func (h *YourController) YourMethod(c *gin.Context) {
    // Implementation
}
```

### Step 2: Add to Main Controller

Update `controller/controller.go`:

```go
type Controller struct {
    User           *UserController
    Role           *RoleController
    Infrastructure *InfrastructureController
    Your           *YourController  // Add your controller
}

func NewController(ctx context.Context, cfg *models.Config, log logger.Logger) *Controller {
    // ... existing initialization ...
    
    yourService := services.NewYourService(/* dependencies */)
    
    return &Controller{
        User:           NewUserController(ctx, userRepo, log, jwtManager),
        Role:           NewRoleController(ctx, roleService, log),
        Infrastructure: NewInfrastructureController(ctx, infraService, log),
        Your:           NewYourController(ctx, yourService, log),  // Initialize
    }
}
```

### Step 3: Register Routes

Add routes in the `RegisterRoutes` method:

```go
func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
    // ... existing route setup ...
    
    // Your routes
    yourGroup := v1.Group("/your-resource")
    {
        yourGroup.GET("/", c.Your.GetAll)
        yourGroup.POST("/", c.Your.jwtManager.AuthMiddleware(), c.Your.Create)
        yourGroup.GET("/:id", c.Your.GetByID)
        yourGroup.PUT("/:id", c.Your.jwtManager.AuthMiddleware(), c.Your.Update)
        yourGroup.DELETE("/:id", c.Your.jwtManager.AuthMiddleware(), c.Your.Delete)
    }
    
    // ... rest of the method
}
```

### Step 4: Add Swagger Documentation

Add Swagger comments to your methods:

```go
// GetAll handles GET /api/v1/your-resource
// @Summary Get all your resources
// @Description Retrieve a list of all your resources
// @Tags Your Resource
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse "Resources retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /your-resource [get]
func (h *YourController) GetAll(c *gin.Context) {
    // Implementation
}
```

## Request/Response Patterns

### Standard Request Handling

```go
func (h *Controller) HandleRequest(c *gin.Context) {
    // 1. Parse and validate request
    var req models.RequestModel
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.APIResponse{
            Status:  "error",
            Code:    http.StatusBadRequest,
            Message: "Invalid request",
            Error: &models.APIError{
                Type:    "ValidationError",
                Details: err.Error(),
            },
        })
        return
    }
    
    // 2. Extract path/query parameters
    id := c.Param("id")
    page := c.DefaultQuery("page", "1")
    
    // 3. Get user context (if authenticated)
    claims, exists := c.Get("jwt_claims")
    if exists {
        jwtClaims := claims.(*models.JWTClaims)
        // Use claims as needed
    }
    
    // 4. Call service layer
    result, err := h.service.ProcessRequest(&req)
    if err != nil {
        // Handle service errors
        c.JSON(http.StatusInternalServerError, models.APIResponse{
            Status:  "error",
            Code:    http.StatusInternalServerError,
            Message: "Processing failed",
            Error: &models.APIError{
                Type:    "ServiceError",
                Details: err.Error(),
            },
        })
        return
    }
    
    // 5. Return success response
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Request processed successfully",
        Data:    result,
    })
}
```

### Pagination Pattern

```go
func (h *Controller) GetWithPagination(c *gin.Context) {
    // Parse pagination parameters
    page := 1
    limit := 10
    
    if pageParam := c.Query("page"); pageParam != "" {
        if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
            page = p
        }
    }
    
    if limitParam := c.Query("limit"); limitParam != "" {
        if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }
    
    // Get data
    allItems, err := h.service.GetAll()
    if err != nil {
        // Handle error
        return
    }
    
    // Apply pagination
    total := len(allItems)
    totalPages := (total + limit - 1) / limit
    offset := (page - 1) * limit
    
    var paginatedItems []interface{}
    if offset < total {
        end := offset + limit
        if end > total {
            end = total
        }
        paginatedItems = allItems[offset:end]
    }
    
    responseData := map[string]interface{}{
        "items": paginatedItems,
        "pagination": map[string]interface{}{
            "page":         page,
            "limit":        limit,
            "total":        total,
            "total_pages":  totalPages,
            "has_next":     page < totalPages,
            "has_previous": page > 1,
        },
    }
    
    c.JSON(http.StatusOK, models.APIResponse{
        Status:  "success",
        Code:    http.StatusOK,
        Message: "Items retrieved successfully",
        Data:    responseData,
    })
}
```

## Error Handling

### Standard Error Response Structure

```go
type APIResponse struct {
    Status  string    `json:"status"`
    Code    int       `json:"code"`
    Message string    `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError `json:"error,omitempty"`
}

type APIError struct {
    Type    string `json:"type"`
    Details string `json:"details"`
}
```

### Error Types and Status Codes

```go
// Validation errors
c.JSON(http.StatusBadRequest, models.APIResponse{
    Status:  "error",
    Code:    http.StatusBadRequest,
    Message: "Invalid request",
    Error: &models.APIError{
        Type:    "ValidationError",
        Details: err.Error(),
    },
})

// Authentication errors
c.JSON(http.StatusUnauthorized, models.APIResponse{
    Status:  "error",
    Code:    http.StatusUnauthorized,
    Message: "Authentication required",
    Error: &models.APIError{
        Type:    "AuthenticationError",
        Details: "User not authenticated",
    },
})

// Authorization errors
c.JSON(http.StatusForbidden, models.APIResponse{
    Status:  "error",
    Code:    http.StatusForbidden,
    Message: "Insufficient permissions",
    Error: &models.APIError{
        Type:    "AuthorizationError",
        Details: "User lacks required permissions",
    },
})

// Not found errors
c.JSON(http.StatusNotFound, models.APIResponse{
    Status:  "error",
    Code:    http.StatusNotFound,
    Message: "Resource not found",
})

// Conflict errors
c.JSON(http.StatusConflict, models.APIResponse{
    Status:  "error",
    Code:    http.StatusConflict,
    Message: "Resource conflict",
    Error: &models.APIError{
        Type:    "ConflictError",
        Details: "Resource already exists",
    },
})

// Server errors
c.JSON(http.StatusInternalServerError, models.APIResponse{
    Status:  "error",
    Code:    http.StatusInternalServerError,
    Message: "Internal server error",
    Error: &models.APIError{
        Type:    "ServerError",
        Details: err.Error(),
    },
})
```

## Best Practices

### 1. Controller Responsibilities
- **Do**: Handle HTTP concerns (parsing, validation, response formatting)
- **Don't**: Implement business logic or database operations

### 2. Error Handling
- Always return structured error responses
- Log errors appropriately based on severity
- Use appropriate HTTP status codes
- Include helpful error messages for clients

### 3. Validation
- Validate input at the controller level
- Use struct tags for basic validation
- Implement custom validation for complex rules
- Return clear validation error messages

### 4. Authentication & Authorization
- Use middleware for authentication checks
- Implement fine-grained permissions where needed
- Extract user context from JWT claims
- Clear permission caches when roles change

### 5. Logging
- Log important operations and errors
- Include relevant context (user ID, request ID, etc.)
- Use structured logging for better searchability

### 6. Response Consistency
- Use consistent response structure across all endpoints
- Include appropriate metadata (pagination, timestamps, etc.)
- Follow RESTful conventions for status codes

### 7. Documentation
- Add Swagger comments for all endpoints
- Include example requests and responses
- Document error conditions and status codes
- Keep documentation up to date with implementation

---

**Related Documentation**: [Services](SERVICES.md) | [Models](MODELS.md) | [Middleware](MIDDLEWARE.md) | [API Reference](API.md)