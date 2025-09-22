package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/services"
	"fieldfuze-backend/utils/logger"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type RoleController struct {
	ctx         context.Context
	roleService *services.RoleService
	logger      logger.Logger
	validator   *validator.Validate
}

func NewRoleController(ctx context.Context, roleService *services.RoleService, logger logger.Logger) *RoleController {
	return &RoleController{
		ctx:         ctx,
		roleService: roleService,
		logger:      logger,
		validator:   validator.New(),
	}
}

// formatValidationErrors formats validation errors into readable messages
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
			case "alpha_unicode":
				errorMessages = append(errorMessages, fieldError.Field()+" must contain only letters, numbers, and unicode characters")
			case "oneof":
				errorMessages = append(errorMessages, fieldError.Field()+" must be one of: "+strings.ReplaceAll(fieldError.Param(), " ", ", "))
			default:
				errorMessages = append(errorMessages, fieldError.Field()+" is invalid")
			}
		}
	}

	return strings.Join(errorMessages, "; ")
}

// GetRoles handles GET /api/v1/auth/user/role
// @Summary Get all roles
// @Description Retrieve a list of all roles
// @Tags Role Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of roles per page"
// @Param status query string false "Filter by role status (active, inactive, archived)"
// @Success 200 {object} models.APIResponse "Roles retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve roles"
// @Router /user/role [get]
func (h *RoleController) GetRoles(c *gin.Context) {
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

	var roles []*models.RoleAssignment
	var err error

	if status != "" {
		roles, err = h.roleService.GetRoleAssignmentsByStatus(status)
	} else {
		roles, err = h.roleService.GetRoleAssignments()
	}

	if err != nil {
		h.logger.Error("Failed to get roles", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get roles",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

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
	} else {
		paginatedRoles = []*models.RoleAssignment{}
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

// CreateRole handles POST /api/v1/auth/user/role
// @Summary Create a new role
// @Description Create a new role with specified permissions
// @Tags Role Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.RoleAssignment true "Create role assignment request"
// @Success 201 {object} models.APIResponse "Role created successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid role data"
// @Failure 409 {object} models.APIResponse "Conflict - Role already exists"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Role creation failed"
// @Router /user/role [post]
func (h *RoleController) CreateRole(c *gin.Context) {
	var req models.RoleAssignment
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

	// Perform struct-level validation
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed:", err)
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

	role, err := h.roleService.CreateRole(h.ctx, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to create role", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "role with this name already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to create role",
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
		Message: "Role created successfully",
		Data:    role,
	})
}

// GetRole handles GET /api/v1/auth/user/role/:id
// @Summary Get role by ID
// @Description Retrieve role details by ID
// @Tags Role Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} models.APIResponse "Role details retrieved successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid role ID"
// @Failure 404 {object} models.APIResponse "Not Found - Role does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve role"
// @Router /user/role/{id} [get]
func (h *RoleController) GetRole(c *gin.Context) {
	roleID := c.Param("id")

	role, err := h.roleService.GetRoleAssignmentByID(roleID)
	if err != nil {
		h.logger.Error("Failed to get role by ID", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" || err.Error() == "role ID is required" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to get role",
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
		Message: "Role details retrieved successfully",
		Data:    role,
	})
}

// UpdateRole handles PUT /api/v1/auth/user/role/:id
// @Summary Update role by ID
// @Description Update role information by ID
// @Tags Role Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body models.RoleAssignment true "Update role assignment request"
// @Success 200 {object} models.APIResponse "Role updated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid role ID or data"
// @Failure 404 {object} models.APIResponse "Not Found - Role does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to update role"
// @Router /user/role/{id} [put]
func (h *RoleController) UpdateRole(c *gin.Context) {
	var req models.RoleAssignment
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

	// Perform struct-level validation
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed:", err)
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

	roleID := c.Param("id")
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

	updatedRole, err := h.roleService.UpdateRoleAssignment(roleID, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to update role", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "role with this name already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to update role",
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
		Message: "Role updated successfully",
		Data:    updatedRole,
	})
}

// DeleteRole handles DELETE /api/v1/auth/user/role/:id
// @Summary Delete role by ID
// @Description Delete role by ID
// @Tags Role Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} models.APIResponse "Role deleted successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid role ID"
// @Failure 404 {object} models.APIResponse "Not Found - Role does not exist"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to delete role"
// @Router /user/role/{id} [delete]
func (h *RoleController) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	err := h.roleService.DeleteRoleAssignment(roleID)
	if err != nil {
		h.logger.Error("Failed to delete role", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "role not found" || err.Error() == "role ID is required" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to delete role",
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
		Message: "Role deleted successfully",
		Data: map[string]interface{}{
			"deleted_role_id": roleID,
		},
	})
}
