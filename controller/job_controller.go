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

type JobController struct {
	ctx        context.Context
	jobService services.JobServiceInterface
	logger     logger.Logger
	validator  *validator.Validate
}

func NewJobController(ctx context.Context, jobService services.JobServiceInterface, logger logger.Logger) *JobController {
	return &JobController{
		ctx:        ctx,
		jobService: jobService,
		logger:     logger,
		validator:  validator.New(),
	}
}

func (h *JobController) formatValidationErrors(err error) string {
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
			case "oneof":
				errorMessages = append(errorMessages, fieldError.Field()+" must be one of: "+strings.ReplaceAll(fieldError.Param(), " ", ", "))
			default:
				errorMessages = append(errorMessages, fieldError.Field()+" is invalid")
			}
		}
	}

	return strings.Join(errorMessages, "; ")
}

// CreateJob handles POST /api/v1/jobs
// @Summary Create a new job
// @Description Create a new job with specified details
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.CreateJobRequest true "Create job request"
// @Success 201 {object} models.APIResponse "Job created successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid job data"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Job creation failed"
// @Router /jobs [post]
func (h *JobController) CreateJob(c *gin.Context) {
	var req models.CreateJobRequest
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

	job, err := h.jobService.CreateJob(h.ctx, &req, jwtClaims.UserID)
	if err != nil {
		h.logger.Error("Failed to create job", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to create job",
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
		Message: "Job created successfully",
		Data:    job,
	})
}

// GetJobs handles GET /api/v1/jobs
// @Summary Get jobs with optional filtering
// @Description Retrieve a list of jobs with optional filtering
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of jobs per page"
// @Param orgID query string false "Filter by organization ID"
// @Param clientID query string false "Filter by client ID"
// @Param jobStatus query string false "Filter by job status"
// @Param jobType query string false "Filter by job type"
// @Param createdBy query string false "Filter by creator"
// @Param fromDate query string false "Filter from date (YYYY-MM-DD)"
// @Param toDate query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} models.APIResponse "Jobs retrieved successfully"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve jobs"
// @Router /jobs [get]
func (h *JobController) GetJobs(c *gin.Context) {
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

	filter := &models.JobFilter{
		OrgID:     c.Query("orgID"),
		ClientID:  c.Query("clientID"),
		CreatedBy: c.Query("createdBy"),
	}

	if jobStatus := c.Query("jobStatus"); jobStatus != "" {
		filter.JobStatus = models.JobStatus(jobStatus)
	}

	if jobType := c.Query("jobType"); jobType != "" {
		filter.JobType = models.JobType(jobType)
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

	jobs, err := h.jobService.GetJobs(filter)
	if err != nil {
		h.logger.Error("Failed to get jobs", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get jobs",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	total := len(jobs)
	totalPages := (total + limit - 1) / limit
	offset := (page - 1) * limit

	var paginatedJobs []*models.Job
	if offset < total {
		end := offset + limit
		if end > total {
			end = total
		}
		paginatedJobs = jobs[offset:end]
	} else {
		paginatedJobs = []*models.Job{}
	}

	responseData := map[string]interface{}{
		"jobs": paginatedJobs,
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
		Message: "Jobs retrieved successfully",
		Data:    responseData,
	})
}

// GetJobByID handles GET /api/v1/jobs/{id}
// @Summary Get job by ID
// @Description Get a specific job by its ID
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.APIResponse "Job retrieved successfully"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id} [get]
func (h *JobController) GetJobByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
			},
		})
		return
	}

	job, err := h.jobService.GetJobByID(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to get job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to get job",
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
		Message: "Job retrieved successfully",
		Data:    job,
	})
}

// UpdateJob handles PUT /api/v1/jobs/{id}
// @Summary Update job
// @Description Update an existing job
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body models.UpdateJobRequest true "Update job request"
// @Success 200 {object} models.APIResponse "Job updated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id} [put]
func (h *JobController) UpdateJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
			},
		})
		return
	}

	var req models.UpdateJobRequest
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

	job, err := h.jobService.UpdateJob(h.ctx, id, &req, jwtClaims.UserID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to update job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to update job",
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
		Message: "Job updated successfully",
		Data:    job,
	})
}

// DeleteJob handles DELETE /api/v1/jobs/{id}
// @Summary Delete job
// @Description Delete a job by ID
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.APIResponse "Job deleted successfully"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id} [delete]
func (h *JobController) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
			},
		})
		return
	}

	err := h.jobService.DeleteJob(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		}
		h.logger.Error("Failed to delete job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to delete job",
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
		Message: "Job deleted successfully",
	})
}

// StartJob handles POST /api/v1/jobs/{id}/start
// @Summary Start a job
// @Description Start a job by changing its status to in_progress
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.APIResponse "Job started successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id}/start [post]
func (h *JobController) StartJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
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

	job, err := h.jobService.StartJob(h.ctx, id, jwtClaims.UserID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "cannot be started") {
			statusCode = http.StatusBadRequest
		}
		h.logger.Error("Failed to start job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to start job",
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
		Message: "Job started successfully",
		Data:    job,
	})
}

// CompleteJob handles POST /api/v1/jobs/{id}/complete
// @Summary Complete a job
// @Description Complete a job by changing its status to completed
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} models.APIResponse "Job completed successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id}/complete [post]
func (h *JobController) CompleteJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
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

	job, err := h.jobService.CompleteJob(h.ctx, id, jwtClaims.UserID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "must be in progress") {
			statusCode = http.StatusBadRequest
		}
		h.logger.Error("Failed to complete job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to complete job",
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
		Message: "Job completed successfully",
		Data:    job,
	})
}

// CancelJob handles POST /api/v1/jobs/{id}/cancel
// @Summary Cancel a job
// @Description Cancel a job by changing its status to cancelled
// @Tags Job Management
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body object{reason=string} false "Cancel reason"
// @Success 200 {object} models.APIResponse "Job cancelled successfully"
// @Failure 400 {object} models.APIResponse "Bad Request"
// @Failure 404 {object} models.APIResponse "Job not found"
// @Failure 500 {object} models.APIResponse "Internal Server Error"
// @Router /jobs/{id}/cancel [post]
func (h *JobController) CancelJob(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Job ID is required",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Job ID parameter is missing",
			},
		})
		return
	}

	var req struct {
		Reason string `json:"reason,omitempty"`
	}
	c.ShouldBindJSON(&req)

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

	job, err := h.jobService.CancelJob(h.ctx, id, jwtClaims.UserID, req.Reason)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "job not found" {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "cannot be cancelled") {
			statusCode = http.StatusBadRequest
		}
		h.logger.Error("Failed to cancel job", err)
		c.JSON(statusCode, models.APIResponse{
			Status:  "error",
			Code:    statusCode,
			Message: "Failed to cancel job",
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
		Message: "Job cancelled successfully",
		Data:    job,
	})
}