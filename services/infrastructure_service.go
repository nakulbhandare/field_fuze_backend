package services

import (
	"context"
	"encoding/json"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type InfrastructureService struct {
	ctx      context.Context
	dbClient *dal.DynamoDBClient
	logger   logger.Logger
	config   *models.Config
}

func NewInfrastructureService(ctx context.Context, dbClient *dal.DynamoDBClient, logger logger.Logger, config *models.Config) *InfrastructureService {
	return &InfrastructureService{
		ctx:      ctx,
		dbClient: dbClient,
		logger:   logger,
		config:   config,
	}
}

// Worker Management Methods

// getWorkerStatus reads worker status from the status file
func (s *InfrastructureService) getWorkerStatus() (*models.ExecutionResult, error) {
	// Get status file path based on environment
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", s.config.AppEnv)

	data, err := os.ReadFile(statusFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read worker status file: %w", err)
	}

	var result models.ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal worker status: %w", err)
	}

	return &result, nil
}

// GetWorkerStatus returns the current worker status with enhanced context (public method for API)
func (s *InfrastructureService) GetWorkerStatus(ctx context.Context) (*models.ExecutionResult, error) {
	s.logger.Debug("Getting detailed worker status")
	
	result, err := s.getWorkerStatus()
	if err != nil {
		return nil, err
	}
	
	// Enrich status with additional context
	s.enrichStatusWithContext(result)
	
	// Update health indicators
	s.updateHealthIndicators(result)
	
	return result, nil
}

// RestartWorker restarts the infrastructure worker
func (s *InfrastructureService) RestartWorker(ctx context.Context, force bool) (*models.ServiceRestartResult, error) {
	s.logger.Info("Restarting infrastructure worker")

	result := &models.ServiceRestartResult{
		ServiceName: "infrastructure-worker",
		StartTime:   time.Now(),
		Status:      "in_progress",
	}

	// Check current worker status
	workerStatus, err := s.getWorkerStatus()
	if err != nil {
		s.logger.Warn("Could not get current worker status, proceeding with restart", err)
	}

	// If worker is running and force is false, return error
	if !force && workerStatus != nil && workerStatus.Status == "running" {
		result.Status = "failed"
		result.Error = "Worker is currently running. Use force=true to restart anyway"
		result.EndTime = time.Now()
		return result, fmt.Errorf("worker is running")
	}

	// Kill existing worker process if running
	if err := s.killWorkerProcess(); err != nil {
		s.logger.Warn("Failed to kill existing worker process", err)
	}

	// Reset worker status to allow restart
	if err := s.resetWorkerStatus(); err != nil {
		s.logger.Warn("Failed to reset worker status", err)
	}

	// Start new worker instance
	if err := s.startWorkerProcess(ctx); err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.Status = "completed"
	result.EndTime = time.Now()
	result.Output = "Worker restart initiated successfully"

	s.logger.Info("Infrastructure worker restart completed")
	return result, nil
}

// enrichStatusWithContext adds contextual information to the execution result
func (s *InfrastructureService) enrichStatusWithContext(result *models.ExecutionResult) {
	// Add next action guidance
	switch result.Status {
	case models.StatusCreatingTables:
		result.NextAction = "Creating DynamoDB tables - this may take a few minutes"
		result.EstimatedTime = s.durationPtr(5 * time.Minute)
		result.Phase = "Table Creation"
		
	case models.StatusWaitingForTables:
		result.NextAction = "Waiting for DynamoDB tables to become active"
		result.EstimatedTime = s.durationPtr(3 * time.Minute)
		result.Phase = "Table Activation"
		
	case models.StatusCreatingIndexes:
		result.NextAction = "Creating database indexes"
		result.EstimatedTime = s.durationPtr(2 * time.Minute)
		result.Phase = "Index Creation"
		
	case models.StatusWaitingForIndexes:
		result.NextAction = "Waiting for database indexes to become ready"
		result.EstimatedTime = s.durationPtr(1 * time.Minute)
		result.Phase = "Index Activation"
		
	case models.StatusValidating:
		result.NextAction = "Validating infrastructure configuration"
		result.EstimatedTime = s.durationPtr(30 * time.Second)
		result.Phase = "Validation"
		
	case models.StatusFixingIssues:
		result.NextAction = "Fixing detected infrastructure issues"
		result.EstimatedTime = s.durationPtr(2 * time.Minute)
		result.Phase = "Issue Resolution"
		
	case models.StatusRevalidating:
		result.NextAction = "Re-validating infrastructure after fixes"
		result.EstimatedTime = s.durationPtr(1 * time.Minute)
		result.Phase = "Re-validation"
		
	case models.StatusFailed:
		if result.RetryCount < 3 {
			result.NextAction = "Will retry automatically after backoff period"
			result.EstimatedTime = s.durationPtr(time.Duration(result.RetryCount+1) * 2 * time.Minute)
		} else {
			result.NextAction = "Manual intervention required - max retries exceeded"
		}
		result.Phase = "Error Recovery"
		
	case models.StatusRetrying:
		result.NextAction = fmt.Sprintf("Retrying infrastructure setup (attempt %d)", result.RetryCount+1)
		result.Phase = "Retry"
		
	case models.StatusCompleted:
		result.NextAction = "Infrastructure is ready for use"
		result.Phase = "Completed"
		
	case models.StatusInitializing:
		result.NextAction = "Initializing infrastructure worker"
		result.Phase = "Initialization"
		
	case models.StatusRunning:
		result.NextAction = "Infrastructure setup is in progress"
		result.Phase = "Setup"
		
	default:
		result.NextAction = "Monitoring infrastructure status"
		result.Phase = "Monitoring"
	}
	
	// Calculate progress if applicable
	if result.Progress == nil {
		result.Progress = s.calculateProgress(result)
	}
}

// updateHealthIndicators sets health status based on execution state
func (s *InfrastructureService) updateHealthIndicators(result *models.ExecutionResult) {
	switch result.Status {
	case models.StatusCompleted:
		if result.Success {
			result.HealthStatus = "healthy"
		} else {
			result.HealthStatus = "degraded"
		}
		
	case models.StatusCreatingTables, models.StatusWaitingForTables,
		 models.StatusCreatingIndexes, models.StatusWaitingForIndexes,
		 models.StatusValidating, models.StatusInitializing:
		result.HealthStatus = "provisioning"
		
	case models.StatusFailed:
		result.HealthStatus = "unhealthy"
		
	case models.StatusRetrying, models.StatusFixingIssues, models.StatusRevalidating:
		result.HealthStatus = "degraded"
		
	case models.StatusRunning:
		// Check how long it's been running
		runningTime := time.Since(result.StartTime)
		if runningTime > 30*time.Minute {
			result.HealthStatus = "degraded" // Running too long
		} else {
			result.HealthStatus = "provisioning"
		}
		
	default:
		result.HealthStatus = "unknown"
	}
}

// calculateProgress estimates setup progress based on current status
func (s *InfrastructureService) calculateProgress(result *models.ExecutionResult) *models.ProgressInfo {
	totalSteps := 6 // Initialization, Create Tables, Wait Tables, Create Indexes, Validate, Complete
	currentStep := 0
	stepName := "Unknown"
	
	switch result.Status {
	case models.StatusInitializing:
		currentStep = 1
		stepName = "Initializing"
	case models.StatusCreatingTables:
		currentStep = 2
		stepName = "Creating Tables"
	case models.StatusWaitingForTables:
		currentStep = 3
		stepName = "Waiting for Tables"
	case models.StatusCreatingIndexes:
		currentStep = 4
		stepName = "Creating Indexes"
	case models.StatusWaitingForIndexes:
		currentStep = 4
		stepName = "Waiting for Indexes"
	case models.StatusValidating, models.StatusRevalidating:
		currentStep = 5
		stepName = "Validating"
	case models.StatusCompleted:
		currentStep = 6
		stepName = "Completed"
	case models.StatusRunning:
		currentStep = 2 // Assume generic running state
		stepName = "In Progress"
	case models.StatusFixingIssues:
		currentStep = 5
		stepName = "Fixing Issues"
	default:
		currentStep = 1
		stepName = string(result.Status)
	}
	
	percentage := (currentStep * 100) / totalSteps
	if percentage > 100 {
		percentage = 100
	}
	
	return &models.ProgressInfo{
		CurrentStep: currentStep,
		TotalSteps:  totalSteps,
		StepName:    stepName,
		Percentage:  percentage,
	}
}

// durationPtr returns a pointer to the given duration
func (s *InfrastructureService) durationPtr(d time.Duration) *time.Duration {
	return &d
}

// killWorkerProcess kills the existing worker process
func (s *InfrastructureService) killWorkerProcess() error {
	// Get lock file path
	lockFilePath := fmt.Sprintf("/tmp/fieldfuze-infrastructure-%s.lock", s.config.AppEnv)

	// Try to read lock file to get PID
	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		s.logger.Debug("No lock file found, worker may not be running")
		return nil
	}

	// Parse lock file content (simplified - in real implementation you'd parse JSON)
	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		return nil
	}

	// Kill process using system command
	cmd := exec.CommandContext(context.Background(), "kill", "-TERM", pidStr)
	if err := cmd.Run(); err != nil {
		// Try force kill
		cmd = exec.CommandContext(context.Background(), "kill", "-KILL", pidStr)
		return cmd.Run()
	}

	return nil
}

// resetWorkerStatus resets the worker status file
func (s *InfrastructureService) resetWorkerStatus() error {
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", s.config.AppEnv)
	return os.Remove(statusFilePath)
}

// startWorkerProcess starts a new worker process
func (s *InfrastructureService) startWorkerProcess(ctx context.Context) error {
	// This would typically start the worker as a separate process
	// For now, we'll simulate by creating a simple status file
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", s.config.AppEnv)

	initialStatus := &models.ExecutionResult{
		StartTime:      time.Now(),
		Status:         "running",
		Environment:    s.config.AppEnv,
		TablesCreated:  make([]models.TableStatus, 0),
		IndexesCreated: make([]models.IndexStatus, 0),
		Metadata:       make(map[string]interface{}),
		RetryCount:     0,
		Success:        false,
	}

	data, err := json.MarshalIndent(initialStatus, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal initial status: %w", err)
	}

	if err := os.WriteFile(statusFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write initial status: %w", err)
	}

	s.logger.Info("Worker process restart initiated")
	return nil
}

// IsWorkerHealthy checks if worker is in a healthy state
func (s *InfrastructureService) IsWorkerHealthy() (bool, string, error) {
	workerStatus, err := s.getWorkerStatus()
	if err != nil {
		return false, "Cannot read worker status", err
	}

	switch workerStatus.Status {
	case "completed":
		if workerStatus.Success {
			return true, "Worker completed successfully", nil
		}
		return false, "Worker completed with errors", nil
	case "running":
		// Check if running too long
		runningTime := time.Since(workerStatus.StartTime)
		if runningTime > 30*time.Minute {
			return false, "Worker running too long", nil
		}
		return true, "Worker is running normally", nil
	case "failed":
		return false, fmt.Sprintf("Worker failed: %s", workerStatus.ErrorMessage), nil
	case "retrying":
		if workerStatus.RetryCount > 5 {
			return false, "Worker stuck in retry loop", nil
		}
		return false, "Worker is retrying after failure", nil
	default:
		return false, "Worker status unknown", nil
	}
}

// AutoRestartIfNeeded checks if worker needs restart and does it automatically
func (s *InfrastructureService) AutoRestartIfNeeded(ctx context.Context) (*models.ServiceRestartResult, error) {
	healthy, reason, err := s.IsWorkerHealthy()
	if err != nil {
		return nil, fmt.Errorf("failed to check worker health: %w", err)
	}

	if healthy {
		return &models.ServiceRestartResult{
			ServiceName: "infrastructure-worker",
			Status:      "not_needed",
			StartTime:   time.Now(),
			EndTime:     time.Now(),
			Output:      "Worker is healthy, no restart needed",
		}, nil
	}

	s.logger.Warnf("Worker is unhealthy (%s), initiating auto-restart", reason)
	return s.RestartWorker(ctx, true)
}
