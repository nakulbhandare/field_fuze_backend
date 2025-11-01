package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/services"
	"fieldfuze-backend/utils/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CrewController struct {
	ctx         context.Context
	crewService services.CrewServiceInterface
	logger      logger.Logger
	validator   *validator.Validate
}

func NewCrewController(ctx context.Context, crewService services.CrewServiceInterface, logger logger.Logger) *CrewController {
	return &CrewController{
		ctx:         ctx,
		crewService: crewService,
		logger:      logger,
		validator:   validator.New(),
	}
}

func (h *CrewController) formatValidationErrors(err error) string {
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

// CreateCrew handles POST /api/v1/crews
// @Summary Create a new crew
// @Description Create a new crew with specified details
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateCrewRequest true "Create crew request"
// @Success 201 {object} models.APIResponse "Crew created successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid crew data"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Crew creation failed"
// @Router /crews [post]
func (h *CrewController) CreateCrew(c *gin.Context) {
	var req models.CreateCrewRequest
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

	crew, err := h.crewService.CreateCrew(h.ctx, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to create crew", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to create crew",
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
		Message: "Crew created successfully",
		Data:    crew,
	})
}

// GetCrews handles GET /api/v1/crews
// @Summary Get crews with optional filtering
// @Description Retrieve a list of crews with optional filtering
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of crews per page"
// @Param orgID query string false "Filter by organization ID"
// @Param leadTechnicianId query string false "Filter by lead technician ID"
// @Param isActive query boolean false "Filter by active status"
// @Param createdBy query string false "Filter by creator"
// @Param fromDate query string false "Filter from date (YYYY-MM-DD)"
// @Param toDate query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} models.APIResponse "Crews retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve crews"
// @Router /crews [get]
func (h *CrewController) GetCrews(c *gin.Context) {
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

	filter := &models.CrewFilter{
		OrgID:            c.Query("orgID"),
		LeadTechnicianId: c.Query("leadTechnicianId"),
		CreatedBy:        c.Query("createdBy"),
	}

	if isActiveStr := c.Query("isActive"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	if fromDateStr := c.Query("fromDate"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filter.FromDate = fromDate
		}
	}

	if toDateStr := c.Query("toDate"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filter.ToDate = toDate
		}
	}

	crews, err := h.crewService.GetCrews(filter)
	if err != nil {
		h.logger.Error("Failed to get crews", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get crews",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	total := len(crews)
	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	var paginatedCrews []*models.Crew
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedCrews = crews[offset:end]
	} else {
		paginatedCrews = []*models.Crew{}
	}

	responseData := map[string]interface{}{
		"crews": paginatedCrews,
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
		Message: "Crews retrieved successfully",
		Data:    responseData,
	})
}

// GetCrewByID handles GET /api/v1/crews/{id}
// @Summary Get crew by ID
// @Description Get a specific crew by its ID
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Crew ID"
// @Success 200 {object} models.APIResponse "Crew retrieved successfully"
// @Failure 404 {object} models.APIResponse "Crew not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /crews/{id} [get]
func (h *CrewController) GetCrewByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Crew ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Crew ID parameter is missing",
			},
		})
		return
	}

	crew, err := h.crewService.GetCrewByID(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "crew not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to get crew", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to get crew",
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
		Message: "Crew retrieved successfully",
		Data:    crew,
	})
}

// UpdateCrew handles PUT /api/v1/crews/{id}
// @Summary Update crew
// @Description Update an existing crew
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Crew ID"
// @Param request body models.UpdateCrewRequest true "Update crew request"
// @Success 200 {object} models.APIResponse "Crew updated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Crew not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /crews/{id} [put]
func (h *CrewController) UpdateCrew(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Crew ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Crew ID parameter is missing",
			},
		})
		return
	}

	var req models.UpdateCrewRequest
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

	crew, err := h.crewService.UpdateCrew(h.ctx, id, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "crew not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to update crew", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to update crew",
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
		Message: "Crew updated successfully",
		Data:    crew,
	})
}

// DeleteCrew handles DELETE /api/v1/crews/{id}
// @Summary Delete crew
// @Description Delete a crew by ID
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Crew ID"
// @Success 200 {object} models.APIResponse "Crew deleted successfully"
// @Failure 404 {object} models.APIResponse "Crew not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /crews/{id} [delete]
func (h *CrewController) DeleteCrew(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Crew ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Crew ID parameter is missing",
			},
		})
		return
	}

	err := h.crewService.DeleteCrew(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "crew not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to delete crew", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to delete crew",
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
		Message: "Crew deleted successfully",
	})
}

// AddMemberToCrew handles POST /api/v1/crews/{id}/members/{memberId}
// @Summary Add member to crew
// @Description Add a member to a crew
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Crew ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} models.APIResponse "Member added successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Crew not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /crews/{id}/members/{memberId} [post]
func (h *CrewController) AddMemberToCrew(c *gin.Context) {
	crewID := c.Param("id")
	memberID := c.Param("memberId")

	if crewID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Crew ID and Member ID are required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Both crew ID and member ID parameters are required",
			},
		})
		return
	}

	crew, err := h.crewService.AddMemberToCrew(h.ctx, crewID, memberID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusBadRequest
		}
		h.logger.Error("Failed to add member to crew", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to add member to crew",
			Error: &models.APIError{
				Type:    "BusinessError",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Member added to crew successfully",
		Data:    crew,
	})
}

// RemoveMemberFromCrew handles DELETE /api/v1/crews/{id}/members/{memberId}
// @Summary Remove member from crew
// @Description Remove a member from a crew
// @Tags Crew Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Crew ID"
// @Param memberId path string true "Member ID"
// @Success 200 {object} models.APIResponse "Member removed successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Crew or member not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /crews/{id}/members/{memberId} [delete]
func (h *CrewController) RemoveMemberFromCrew(c *gin.Context) {
	crewID := c.Param("id")
	memberID := c.Param("memberId")

	if crewID == "" || memberID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Crew ID and Member ID are required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Both crew ID and member ID parameters are required",
			},
		})
		return
	}

	crew, err := h.crewService.RemoveMemberFromCrew(h.ctx, crewID, memberID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to remove member from crew", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to remove member from crew",
			Error: &models.APIError{
				Type:    "BusinessError",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Member removed from crew successfully",
		Data:    crew,
	})
}