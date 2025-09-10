package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fmt"
	"net/http"
	"strconv"
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

	users, err := h.userRepo.GetUser(userID)
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

// GetUserList handles GET /api/v1/auth/user/list
// @Summary Get list of users
// @Description Retrieve a list of all users
// @Tags User Management
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of users per page"
// @Param sort query string false "Sort order (e.g., 'asc' or 'desc')"
// @Success 200 {object} models.APIResponse "User list retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve user list"
// @Router /auth/user/list [get]
func (h *UserController) GetUserList(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10
	sort := "asc"

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

	if sortParam := c.Query("sort"); sortParam == "desc" {
		sort = "desc"
	}

	// Get all users
	allUsers, err := h.userRepo.GetUser("")
	if err != nil {
		h.logger.Error("Failed to get user list", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user list",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	// Sort users by created_at
	if sort == "desc" {
		for i, j := 0, len(allUsers)-1; i < j; i, j = i+1, j-1 {
			allUsers[i], allUsers[j] = allUsers[j], allUsers[i]
		}
	}

	// Calculate pagination
	total := len(allUsers)
	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	// Apply pagination
	var paginatedUsers []*models.User
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedUsers = allUsers[offset:end]
	} else {
		paginatedUsers = []*models.User{}
	}

	// Create response with pagination metadata
	responseData := map[string]interface{}{
		"users": paginatedUsers,
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
		Message: "User list retrieved successfully",
		Data:    responseData,
	})
}

// UpdateUser handles PATCH /api/v1/auth/user/update/{id}
// @Summary Update user details
// @Description Update user information by ID
// @Tags User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.User true "Update user request"
// @Success 200 {object} models.APIResponse "User updated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID or data"
// @Failure 404 {object} models.APIResponse "Not Found - User does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to update user"
// @Router /auth/user/update/{id} [patch]
func (h *UserController) UpdateUser(c *gin.Context) {
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

	userID := c.Param("id")
	if userID == "" {
		h.logger.Error("Missing user ID")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Missing user ID",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "User ID is required",
			},
		})
		return
	}

	// Update user in the repository
	updatedUser, err := h.userRepo.UpdateUser(userID, &req)
	if err != nil {
		h.logger.Error("Failed to update user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to update user",
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
		Message: "User updated successfully",
		Data:    updatedUser,
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

	users, err := h.userRepo.GetUser(email)
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

	if len(users) == 0 {
		h.logger.Error("User not found")
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

	user := users[0]
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
