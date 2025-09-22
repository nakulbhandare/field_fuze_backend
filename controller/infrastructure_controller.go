package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/services"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type InfrastructureController struct {
	ctx     context.Context
	service *services.InfrastructureService
	logger  logger.Logger
}

func NewInfrastructureController(ctx context.Context, service *services.InfrastructureService, logger logger.Logger) *InfrastructureController {
	return &InfrastructureController{
		ctx:     ctx,
		service: service,
		logger:  logger,
	}
}

// GetWorkerStatus handles GET /api/v1/infrastructure/worker/status
// @Summary Get worker execution status
// @Description Retrieve detailed status of the infrastructure worker including execution state, progress, and health
// @Tags Infrastructure
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse "Worker status retrieved successfully"
// @Failure 401 {object} models.APIResponse "Unauthorized - Authentication required"
// @Failure 403 {object} models.APIResponse "Forbidden - Admin access required"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to retrieve worker status"
// @Router /infrastructure/worker/status [get]
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

	// Map worker execution status to appropriate HTTP status
	httpStatus, apiStatus := h.mapWorkerStatusToHTTP(workerStatus)
	message := h.getStatusMessage(workerStatus)
	
	c.JSON(httpStatus, models.APIResponse{
		Status:  apiStatus,
		Code:    httpStatus,
		Message: message,
		Data:    workerStatus,
	})
}

// RestartWorker handles POST /api/v1/infrastructure/worker/restart
// @Summary Restart infrastructure worker
// @Description Restart the infrastructure worker with optional force parameter to restart even if currently running
// @Tags Infrastructure
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.WorkerRestartRequest false "Worker restart options"
// @Success 200 {object} models.APIResponse "Worker restart initiated successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid restart request"
// @Failure 401 {object} models.APIResponse "Unauthorized - Authentication required"
// @Failure 403 {object} models.APIResponse "Forbidden - Admin access required"
// @Failure 409 {object} models.APIResponse "Conflict - Worker is running and force=false"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to restart worker"
// @Router /infrastructure/worker/restart [post]
func (h *InfrastructureController) RestartWorker(c *gin.Context) {
	var restartRequest models.WorkerRestartRequest
	if err := c.ShouldBindJSON(&restartRequest); err != nil {
		// If no body provided, use defaults
		restartRequest = models.WorkerRestartRequest{
			Force: false,
		}
	}

	result, err := h.service.RestartWorker(h.ctx, restartRequest.Force)
	if err != nil {
		// Check if it's a conflict (worker running)
		if strings.Contains(err.Error(), "worker is running") {
			h.logger.Warnf("Worker restart denied - worker is currently running")
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

		h.logger.Errorf("Failed to restart worker: %v", err)
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

	h.logger.Info("Worker restart initiated successfully")
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Worker restart initiated successfully",
		Data:    result,
	})
}

// CheckWorkerHealth handles GET /api/v1/infrastructure/worker/health
// @Summary Check worker health
// @Description Check if the infrastructure worker is healthy and get health details
// @Tags Infrastructure
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse "Worker health check completed"
// @Failure 401 {object} models.APIResponse "Unauthorized - Authentication required"
// @Failure 403 {object} models.APIResponse "Forbidden - Admin access required"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to check worker health"
// @Router /infrastructure/worker/health [get]
func (h *InfrastructureController) CheckWorkerHealth(c *gin.Context) {
	healthy, reason, err := h.service.IsWorkerHealthy()
	if err != nil {
		h.logger.Error("Failed to check worker health", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to check worker health",
			Error: &models.APIError{
				Type:    "WorkerError",
				Details: err.Error(),
			},
		})
		return
	}

	healthStatus := "healthy"
	if !healthy {
		healthStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Worker health check completed",
		Data: map[string]interface{}{
			"healthy": healthy,
			"status":  healthStatus,
			"reason":  reason,
		},
	})
}

// AutoRestartWorker handles POST /api/v1/infrastructure/worker/auto-restart
// @Summary Auto-restart worker if needed
// @Description Check worker health and automatically restart if unhealthy
// @Tags Infrastructure
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse "Auto-restart check completed"
// @Failure 401 {object} models.APIResponse "Unauthorized - Authentication required"
// @Failure 403 {object} models.APIResponse "Forbidden - Admin access required"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Failed to auto-restart worker"
// @Router /infrastructure/worker/auto-restart [post]
func (h *InfrastructureController) AutoRestartWorker(c *gin.Context) {
	result, err := h.service.AutoRestartIfNeeded(h.ctx)
	if err != nil {
		h.logger.Error("Failed to auto-restart worker", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to auto-restart worker",
			Error: &models.APIError{
				Type:    "WorkerError",
				Details: err.Error(),
			},
		})
		return
	}

	message := "Auto-restart check completed"
	if result.Status == "not_needed" {
		message = "Worker is healthy, no restart needed"
	} else if result.Status == "completed" {
		message = "Worker was unhealthy and has been restarted"
	}

	h.logger.Infof("Auto-restart check completed: %s", result.Status)
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: message,
		Data:    result,
	})
}

// mapWorkerStatusToHTTP maps worker execution status to appropriate HTTP status codes
func (h *InfrastructureController) mapWorkerStatusToHTTP(ws *models.ExecutionResult) (int, string) {
	switch ws.Status {
	case models.StatusCompleted:
		if ws.Success {
			return http.StatusOK, "success"
		}
		return http.StatusOK, "warning" // Completed but with issues
		
	case models.StatusFailed:
		return http.StatusServiceUnavailable, "error" // Infrastructure not ready
		
	case models.StatusCreatingTables, models.StatusWaitingForTables,
		 models.StatusCreatingIndexes, models.StatusWaitingForIndexes,
		 models.StatusValidating, models.StatusFixingIssues, models.StatusRevalidating,
		 models.StatusInitializing, models.StatusRunning:
		return http.StatusAccepted, "in_progress" // 202 - Accepted, processing
		
	case models.StatusRetrying:
		return http.StatusAccepted, "retrying"
		
	case models.StatusDeleting, models.StatusDeletionScheduled:
		return http.StatusAccepted, "deleting"
		
	case models.StatusDeleted:
		return http.StatusOK, "deleted"
		
	case models.StatusDeletionFailed:
		return http.StatusServiceUnavailable, "deletion_failed"
		
	default:
		return http.StatusOK, "info"
	}
}

// getStatusMessage provides human-readable status messages
func (h *InfrastructureController) getStatusMessage(ws *models.ExecutionResult) string {
	switch ws.Status {
	case models.StatusCompleted:
		if ws.Success {
			return "Infrastructure is ready and healthy"
		}
		return "Infrastructure setup completed with warnings"
		
	case models.StatusFailed:
		return "Infrastructure setup failed - manual intervention may be required"
		
	case models.StatusCreatingTables:
		return "Creating DynamoDB tables"
		
	case models.StatusWaitingForTables:
		return "Waiting for DynamoDB tables to become active"
		
	case models.StatusCreatingIndexes:
		return "Creating database indexes"
		
	case models.StatusWaitingForIndexes:
		return "Waiting for database indexes to become ready"
		
	case models.StatusValidating:
		return "Validating infrastructure configuration"
		
	case models.StatusFixingIssues:
		return "Fixing detected infrastructure issues"
		
	case models.StatusRevalidating:
		return "Re-validating infrastructure after fixes"
		
	case models.StatusRetrying:
		return fmt.Sprintf("Retrying infrastructure setup (attempt %d)", ws.RetryCount+1)
		
	case models.StatusInitializing:
		return "Initializing infrastructure worker"
		
	case models.StatusRunning:
		return "Infrastructure setup is running"
		
	case models.StatusDeleting:
		return "Deleting infrastructure resources"
		
	case models.StatusDeletionScheduled:
		return "Infrastructure deletion has been scheduled"
		
	case models.StatusDeleted:
		return "Infrastructure has been successfully deleted"
		
	case models.StatusDeletionFailed:
		return "Infrastructure deletion failed"
		
	default:
		return "Worker status retrieved successfully"
	}
}
