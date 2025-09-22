package controller

import (
	"context"
	"fieldfuze-backend/middelware"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fmt"
	"net/http"
	"strconv"

	"fieldfuze-backend/utils/logger"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	ctx        context.Context
	userRepo   *repository.UserRepository
	jwtManager *middelware.JWTManager
	logger     logger.Logger
}

func NewUserController(ctx context.Context, userRepo *repository.UserRepository, logger logger.Logger, jwtManager *middelware.JWTManager) *UserController {
	return &UserController{
		ctx:        ctx,
		userRepo:   userRepo,
		logger:     logger,
		jwtManager: jwtManager,
	}
}

// invalidateUserPermissions clears permission cache and logs security events
func (h *UserController) invalidateUserPermissions(userID, operation string) {
	// Clear permission cache through JWT manager
	h.jwtManager.ClearPermissionCache()

	// Log security event for audit trail
	h.logger.Infof("SECURITY EVENT: Permission cache cleared for user %s due to %s", userID, operation)
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
// @Router /user/register [post]
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
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.APIResponse "User details retrieved successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID"
// @Failure 404 {object} models.APIResponse "Not Found - User does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve user"
// @Router /user/{id} [get]
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
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of users per page"
// @Param sort query string false "Sort order (e.g., 'asc' or 'desc')"
// @Success 200 {object} models.APIResponse "User list retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve user list"
// @Router /user/list [get]
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
// @Description Update user information by ID. Note: Role and Roles fields are ignored - use dedicated role assignment endpoints instead.
// @Tags User Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.User true "Update user request (role/roles fields will be ignored)"
// @Success 200 {object} models.APIResponse "User updated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID or data"
// @Failure 404 {object} models.APIResponse "Not Found - User does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to update user"
// @Router /user/update/{id} [patch]
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

	// Explicitly clear role fields to prevent updates
	req.Role = ""
	req.Roles = nil

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

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

// Login handles POST /api/v1/auth/user/login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} models.APIResponse "Login successful, returns JWT token"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid login data"
// @Failure 401 {object} models.APIResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Login failed"
// @Router /user/login [post]
func (h *UserController) Login(c *gin.Context) {
	// Delegate to the JWT manager's login authentication handler
	h.jwtManager.HandleLogin(c)
}

// GenerateToken handles POST /api/v1/auth/user/token
// @Summary Generate JWT token
// @Description Generate or refresh JWT token (legacy endpoint - use /login instead)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.User true "Token generation request"
// @Success 200 {object} models.APIResponse "Token generated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid token request"
// @Failure 401 {object} models.APIResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Token generation failed"
// @Router /user/token [POST]
// //
func (h *UserController) GenerateToken(c *gin.Context) {
	// This endpoint is handled entirely by the AuthMiddleware
	// The middleware detects login requests and processes them automatically
}

// Logout handles POST /api/v1/auth/user/logout
// @Summary User logout
// @Description Logout user and revoke current JWT token
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse "Logout successful"
// @Failure 401 {object} models.APIResponse "Unauthorized - Invalid or missing token"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Logout failed"
// @Router /user/logout [post]
func (h *UserController) Logout(c *gin.Context) {
	// Extract JWT claims from context (set by auth middleware)
	claims, exists := c.Get("jwt_claims")
	if !exists {
		h.logger.Error("JWT claims not found in context")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Code:    http.StatusUnauthorized,
			Message: "Authentication required",
			Error: &models.APIError{
				Type:    "AuthenticationError",
				Details: "User not authenticated",
			},
		})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		h.logger.Error("Invalid JWT claims type")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Invalid token claims",
			Error: &models.APIError{
				Type:    "TokenError",
				Details: "Invalid token structure",
			},
		})
		return
	}

	// Revoke the current token (both blacklist and remove from active tokens)
	h.jwtManager.RevokeUserToken(jwtClaims.UserID, jwtClaims.ID, jwtClaims.ExpiresAt.Time)

	h.logger.Debugf("User %s logged out successfully", jwtClaims.UserID)

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Logout successful",
		Data: map[string]interface{}{
			"logged_out_at": jwtClaims.ExpiresAt.Time,
			"user_id":       jwtClaims.UserID,
		},
	})
}

// ValidateToken godoc
// @Summary      Validate JWT token
// @Description  Validate a JWT token and return user information with roles
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body middelware.TokenValidationRequest true "Token validation request"
// @Success      200  {object}  models.APIResponse  "Token is valid"
// @Failure      400  {object}  models.APIResponse  "Bad Request - Missing or invalid token in request body"
// @Failure      401  {object}  models.APIResponse  "Unauthorized - Invalid or expired token"
// @Router       /user/validate [post]
func (h *UserController) ValidateToken(c *gin.Context) {
	// Delegate to JWT middleware which handles the complete token validation flow
	h.jwtManager.ValidateTokenEndpoint(c)
}

// AssignRole handles POST /api/v1/auth/user/{user_id}/role/{role_id}
// @Summary Assign existing role to user
// @Description Assign an existing role by ID to a user
// @Tags User Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 200 {object} models.APIResponse "Role assigned successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID or role ID"
// @Failure 403 {object} models.APIResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} models.APIResponse "Not Found - User or role does not exist"
// @Failure 409 {object} models.APIResponse "Conflict - User already has this role"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to assign role"
// @Router /user/{user_id}/role/{role_id} [post]
func (h *UserController) AssignRole(c *gin.Context) {
	userID := c.Param("user_id")
	roleID := c.Param("role_id")

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

	if roleID == "" {
		h.logger.Error("Missing role ID")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Missing role ID",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Role ID is required",
			},
		})
		return
	}

	// Check if user exists
	users, err := h.userRepo.GetUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user",
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

	// Assign role to user using the existing method
	updatedUser, err := h.userRepo.AssignRoleToUser(h.ctx, userID, roleID)
	if err != nil {
		if err.Error() == "user already has this role" {
			c.JSON(http.StatusConflict, models.APIResponse{
				Status:  "error",
				Code:    http.StatusConflict,
				Message: "User already has this role",
				Error: &models.APIError{
					Type:    "ConflictError",
					Details: "User already has this role assigned",
				},
			})
			return
		}

		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Status:  "error",
				Code:    http.StatusNotFound,
				Message: "Role not found",
				Error: &models.APIError{
					Type:    "NotFoundError",
					Details: "The specified role does not exist",
				},
			})
			return
		}

		h.logger.Error("Failed to assign role to user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to assign role to user",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	// Clear permission cache for the user after role assignment
	h.invalidateUserPermissions(userID, "role_assignment")

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Role assigned successfully",
		Data:    updatedUser,
	})
}

// DetachRole handles DELETE /api/v1/auth/user/{user_id}/role/{role_id}
// @Summary Remove role from user
// @Description Remove an existing role from a user
// @Tags User Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 200 {object} models.APIResponse "Role removed successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid user ID or role ID"
// @Failure 403 {object} models.APIResponse "Forbidden - Insufficient permissions"
// @Failure 404 {object} models.APIResponse "Not Found - User or role does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to remove role"
// @Router /user/{user_id}/role/{role_id} [delete]
func (h *UserController) DetachRole(c *gin.Context) {
	userID := c.Param("user_id")
	roleID := c.Param("role_id")

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

	if roleID == "" {
		h.logger.Error("Missing role ID")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Missing role ID",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Role ID is required",
			},
		})
		return
	}

	// Check if user exists
	users, err := h.userRepo.GetUser(userID)
	if err != nil {
		h.logger.Error("Failed to get user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user",
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

	// Remove role from user using the existing method
	updatedUser, err := h.userRepo.RemoveRoleFromUser(h.ctx, userID, roleID)
	if err != nil {
		if err.Error() == "role not found for user" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Status:  "error",
				Code:    http.StatusNotFound,
				Message: "User does not have this role",
				Error: &models.APIError{
					Type:    "NotFoundError",
					Details: "User does not have this role assigned",
				},
			})
			return
		}

		h.logger.Error("Failed to remove role from user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to remove role from user",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	// Clear permission cache for the user after role removal
	h.invalidateUserPermissions(userID, "role_removal")

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Role removed successfully",
		Data:    updatedUser,
	})
}
