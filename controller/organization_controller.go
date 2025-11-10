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

type OrganizationController struct {
	ctx                 context.Context
	organizationService services.OrganizationServiceInterface
	logger              logger.Logger
	validator           *validator.Validate
}

func NewOrganizationController(ctx context.Context, organizationService services.OrganizationServiceInterface, logger logger.Logger) *OrganizationController {
	return &OrganizationController{
		ctx:                 ctx,
		organizationService: organizationService,
		logger:              logger,
		validator:           validator.New(),
	}
}

// formatValidationErrors formats validation errors into readable messages
func (h *OrganizationController) formatValidationErrors(err error) string {
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

// CreateOrganization handles POST /api/v1/auth/organization
// @Summary Create a new organization
// @Description Create a new organization with specified details
// @Tags Organization Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.Organization true "Create organization request"
// @Success 201 {object} models.APIResponse "Organization created successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid organization data"
// @Failure 409 {object} models.APIResponse "Conflict - Organization already exists"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Organization creation failed"
// @Router /organization [post]
func (h *OrganizationController) CreateOrganization(c *gin.Context) {
	var req models.Organization
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

	organization, err := h.organizationService.CreateOrganization(h.ctx, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to create organization", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "organization with this name already exists" {
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
		Message: "Organization created successfully",
		Data:    organization,
	})

}

// GetOrganizations handles GET /api/v1/auth/organization
// @Summary Get all organizations
// @Description Retrieve a list of all organizations
// @Tags Organization Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of organizations per page"
// @Param status query string false "Filter by organization status (active, inactive, archived)"
// @Success 200 {object} models.APIResponse "Organizations retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve organizations"
// @Router /user/organization [get]
func (h *OrganizationController) GetOrganizations(c *gin.Context) {
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

	var err error

	organizations, err := h.organizationService.GetOrganizations("")

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

	total := len(organizations)
	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	var paginatedOrganizations []*models.Organization
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedOrganizations = organizations[offset:end]
	} else {
		paginatedOrganizations = []*models.Organization{}
	}

	responseData := map[string]interface{}{
		"organizations": paginatedOrganizations,
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

func (h *OrganizationController) UpdateOrganization(c *gin.Context) {
	var req models.Organization
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

	organization, err := h.organizationService.UpdateOrganization(req.ID, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to update organization", err)
		statusCode := http.StatusInternalServerError
		if err.Error() == "organization with this name already exists" {
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

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Organization updated successfully",
		Data:    organization,
	})

}

func (h *OrganizationController) DeleteOrganization(c *gin.Context) {
	organizationID := c.Param("id")
	if organizationID == "" {
		h.logger.Error("Organization ID is required")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Organization ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Organization ID cannot be empty",
			},
		})
		return
	}

	err := h.organizationService.DeleteOrganization(organizationID)
	if err != nil {
		h.logger.Error("Failed to delete organization", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete organization",
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
		Message: "Organization deleted successfully",
	})
}
