package controller

import (
	"context"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/middelware"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"

	"fieldfuze-backend/utils/swagger"
	"net/http"

	"fieldfuze-backend/utils/logger"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	User *UserController
}

func NewController(ctx context.Context, cfg *models.Config, log logger.Logger) *Controller {
	dbclient, err := dal.NewDynamoDBClient(cfg, log)
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB client: %v", err)
	}

	userRepo := repository.NewUserRepository(dbclient, cfg, log)
	jwtManager := middelware.NewJWTManager(cfg, log, userRepo)

	return &Controller{
		User: NewUserController(ctx, userRepo, log, jwtManager),
	}
}

func (c *Controller) RegisterRoutes(ctx context.Context, config *models.Config, r *gin.Engine, basePath string) error {
	v1 := r.Group(basePath)

	// Auth routes
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

	user := v1.Group("/user")

	// User routes - authentication not required
	user.POST("/register", c.User.Register)

	// Login routes - handled by AuthMiddleware
	user.POST("/login", c.User.jwtManager.AuthMiddleware(), c.User.Login)
	user.POST("/token", c.User.jwtManager.AuthMiddleware(), c.User.GenerateToken)
	user.POST("/validate", c.User.ValidateToken)

	// User routes - authentication required
	user.POST("/logout", c.User.jwtManager.AuthMiddleware(), c.User.Logout)
	user.GET("/:id", c.User.jwtManager.AuthMiddleware(), c.User.GetUser)
	user.GET("/list", c.User.jwtManager.AuthMiddleware(), c.User.GetUserList)
	user.PATCH("/update/:id", c.User.jwtManager.AuthMiddleware(), c.User.UpdateUser)

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
