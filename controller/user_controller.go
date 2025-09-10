package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fmt"
	"net/http"
	"time"

	"fieldfuze-backend/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserController struct {
	ctx      context.Context
	userRepo *repository.UserRepository
	// jwtManager *auth.JWTManager
	logger logger.Logger
}

func NewUserController(ctx context.Context, userRepo *repository.UserRepository, logger logger.Logger) *UserController {
	return &UserController{
		ctx:      ctx,
		userRepo: userRepo,
		logger:   logger,
	}
}

// Register handles POST /api/v1/auth/user/register
// @Summary Register a new user
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterUser true "Registration request"
// @Success 201 {object} models.APIResponse "User registered successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid registration data"
// @Failure 409 {object} models.APIResponse "Conflict - User already exists"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Registration failed"
// @Router /auth/user/register [post]
func (h *UserController) Register(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON:", err)
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

	user, err := h.userRepo.CreateUser(h.ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Status:  "success",
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data:    user,
	})
}

// GetUser handles GET /api/v1/auth/user
// @Summary Get user details
// @Description Retrieve user details by ID
// @Tags User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.APIResponse "User details retrieved successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID"
// @Failure 404 {object} models.APIResponse "Not Found - User does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve user"
// @Router /auth/user/{id} [get]
func (h *UserController) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userRepo.GetUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user by ID", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user by ID",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "User details retrieved successfully",
		Data:    user,
	})
}

// GenerateToken handles POST /api/v1/auth/user/token
// @Summary Generate JWT token
// @Description Generate or refresh JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterUser true "Token generation request"
// @Success 200 {object} models.APIResponse "Token generated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid token request"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Token generation failed"
// @Router /auth/user/token [POST]
func (h *UserController) GenerateToken(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON:", err)
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

	email, ok := req.Email, req.Email != ""
	if !ok {
		h.logger.Error("Missing email or password")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Missing email or password",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Email and password are required",
			},
		})
		return
	}

	user, err := h.userRepo.GetUser(email)
	if err != nil {
		h.logger.Error("Failed to get user by email", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user by email",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	if user.Password != req.Password {
		h.logger.Error("Invalid password")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Code:    http.StatusUnauthorized,
			Message: "Invalid email or password",
			Error: &models.APIError{
				Type:    "AuthenticationError",
				Details: "Invalid email or password",
			},
		})
		return
	}

	roles := models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "User",
		Permissions: []string{"read", "write"},
		Level:       1,
		Context: map[string]string{
			"project_id": "project-123",
			"org_id":     "org-123",
		},
		AssignedAt: time.Now(),
		ExpiresAt:  nil,
	}

	// Create JWT claims
	claims := models.JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Status:   user.Status,
		Roles:    []models.RoleAssignment{roles}, // Simplified for this example
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "FieldFuze",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // token valid for 1 hour
		},
		Context: models.UserContext{
			OrganizationID: "org-123",
			CustomerID:     "cust-123",
			WorkerID:       "worker-123",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token generation failed",
			"message": err.Error(),
		})
		return
	}

	fmt.Println("Generated Token:", tokenString)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}
