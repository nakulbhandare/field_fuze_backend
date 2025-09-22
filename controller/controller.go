package controller

import (
	"context"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/middelware"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/services"

	"fieldfuze-backend/utils/swagger"
	"net/http"

	"fieldfuze-backend/utils/logger"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	User           *UserController
	Role           *RoleController
	Infrastructure *InfrastructureController
}

func NewController(ctx context.Context, cfg *models.Config, log logger.Logger) *Controller {
	dbclient, err := dal.NewDynamoDBClient(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	userRepo := repository.NewUserRepository(dbclient, cfg, log)
	roleRepo := repository.NewRoleRepository(dbclient, cfg, log)
	jwtManager := middelware.NewJWTManager(cfg, log, userRepo)

	roleService := services.NewRoleService(roleRepo, log)
	infraService := services.NewInfrastructureService(ctx, dbclient, log, cfg)

	return &Controller{
		User:           NewUserController(ctx, userRepo, log, jwtManager),
		Role:           NewRoleController(ctx, roleService, log),
		Infrastructure: NewInfrastructureController(ctx, infraService, log),
	}
}

func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
	// Apply CORS middleware globally
	corsMiddleware := middelware.NewCORSMiddleware(config)
	r.Use(corsMiddleware.CORS())

	// Add request logging middleware
	loggingMiddleware := middelware.NewLoggingMiddleware(logger.NewLogger(config.LogLevel, config.LogFormat))
	r.Use(loggingMiddleware.StructuredLogger())
	r.Use(loggingMiddleware.Recovery())

	v1 := r.Group(basePath)

	// Health check endpoint (no auth required)
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"version": "1.0.0",
			"service": "FieldFuze Backend",
		})
	})

	// Serve static assets for Swagger UI (fallback to embedded swagger assets)
	r.Static("/swagger-ui-assets", "./node_modules/swagger-ui-dist")

	// Swagger UI with authentication form
	swaggerConfig := swagger.SwaggerConfig{
		Title:         "FieldFuze Backend API",
		SwaggerDocURL: "/swagger/doc.json",
		AuthURL:       "/api/v1/auth/user/login",
	}
	r.GET("/swagger", swagger.ServeCleanSwagger(swaggerConfig))
	r.GET("/swagger/", swagger.ServeCleanSwagger(swaggerConfig))
	r.GET("/swagger/index.html", swagger.ServeCleanSwagger(swaggerConfig))

	// Swagger JSON spec
	r.GET("/swagger/doc.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.File("./docs/swagger.json")
	})

	// User group (base path already includes /auth)
	user := v1.Group("/user")

	// Public routes - authentication not required
	user.POST("/register", c.User.Register)
	user.POST("/login", c.User.Login)            // No auth needed - users don't have tokens yet
	user.POST("/token", c.User.GenerateToken)    // No auth needed - token generation endpoint
	user.POST("/validate", c.User.ValidateToken) // No auth needed - validates tokens manually

	// Protected routes - authentication + enhanced authorization required
	user.POST("/logout", c.User.jwtManager.AuthMiddleware(), c.User.Logout)
	user.GET("/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_details"), c.User.GetUser)            // Resource-specific: user details with context validation
	user.GET("/list", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_list"), c.User.GetUserList)          // Resource-specific: user list with department scope
	user.PATCH("/update/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("user_update"), c.User.UpdateUser) // Resource-specific: user update with ownership check

	// Role assignment routes - resource-specific permissions with level requirements
	user.POST("/:user_id/role/:role_id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_assign"), c.User.AssignRole)   // Resource-specific: role assignment with level 7+ requirement
	user.DELETE("/:user_id/role/:role_id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_assign"), c.User.DetachRole) // Resource-specific: role assignment with level 7+ requirement

	// Role management routes - resource-specific permissions with context validation
	user.GET("/role", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_list"), c.Role.GetRoles)            // Resource-specific: role list with department scope
	user.POST("/role", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_create"), c.Role.CreateRole)       // Resource-specific: role creation with level 6+ requirement
	user.GET("/role/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_list"), c.Role.GetRole)         // Resource-specific: role details with department scope
	user.PUT("/role/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_update"), c.Role.UpdateRole)    // Resource-specific: role update with level 6+ requirement
	user.DELETE("/role/:id", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequireResourcePermission("role_delete"), c.Role.DeleteRole) // Resource-specific: role deletion with level 8+ requirement

	// Infrastructure routes (require admin permissions)
	infra := v1.Group("/infrastructure", c.User.jwtManager.AuthMiddleware(), c.User.jwtManager.RequirePermission("admin"))
	{

		// Worker-specific management endpoints
		worker := infra.Group("/worker")
		{
			worker.GET("/status", c.Infrastructure.GetWorkerStatus)          // Get worker execution status
			worker.GET("/health", c.Infrastructure.CheckWorkerHealth)        // Check worker health
			worker.POST("/restart", c.Infrastructure.RestartWorker)          // Restart worker
			worker.POST("/auto-restart", c.Infrastructure.AutoRestartWorker) // Auto-restart if unhealthy
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    config.AppHost + ":" + config.AppPort,
		Handler: r,
	}
	// Start server
	logger := logger.NewLogger(config.LogLevel, config.LogFormat)
	logger.Infof("🚀 Starting server on %s:%s", config.AppHost, config.AppPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
