package worker

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"time"
)

// Service wraps the infrastructure worker for easy integration
type Service struct {
	worker *models.Worker
	logger logger.Logger
}

// NewService creates a new worker service
func NewService(ctx context.Context, cfg *models.Config, log logger.Logger) (*Service, error) {
	worker, err := NewWorker(ctx, cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure worker: %w", err)
	}

	return &Service{
		worker: worker,
		logger: log,
	}, nil
}

// StartInBackground starts the infrastructure worker in the background
func (s *Service) StartInBackground() error {
	s.logger.Info("Starting infrastructure worker service in background")

	// Use the models.Worker directly without copying it
	worker := s.worker

	// Start worker in a separate goroutine
	go func() {
		// Create a wrapper that references the original worker (no copy)
		w := &Worker{Worker: worker} // Use pointer, no copying
		if err := w.Start(); err != nil {
			s.logger.Errorf("Infrastructure worker failed to start: %v", err)
		}
	}()

	return nil
}

// Stop stops the infrastructure worker service
func (s *Service) Stop() error {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use
	s.logger.Info("Stopping infrastructure worker service")
	return w.Stop()
}

// GetStatus returns the current infrastructure setup status
func (s *Service) GetStatus() (*models.ExecutionResult, error) {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use
	return w.GetStatus()
}

// IsSetupCompleted checks if infrastructure setup is completed
func (s *Service) IsSetupCompleted() (bool, error) {
	status, err := s.GetStatus()
	if err != nil {
		return false, err
	}

	return status.Status == models.StatusCompleted && status.Success, nil
}

// GetHealthStatus returns a health status for monitoring
func (s *Service) GetHealthStatus() map[string]interface{} {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use pointer, no copying
	status, err := s.GetStatus()
	if err != nil {
		return map[string]interface{}{
			"status":         "error",
			"message":        fmt.Sprintf("Failed to get status: %v", err),
			"healthy":        false,
			"worker_running": w.IsRunning(),
		}
	}

	healthy := status.Status == models.StatusCompleted && status.Success

	// Extract retry count from metadata
	retryCount := 0
	if status.Metadata != nil {
		if rc, ok := status.Metadata["retry_count"].(int); ok {
			retryCount = rc
		}
	}

	return map[string]interface{}{
		"status":         string(status.Status),
		"healthy":        healthy,
		"worker_running": w.IsRunning(),
		"tables_created": status.TablesCreated,
		"retry_count":    retryCount,
		"environment":    status.Environment,
		"start_time":     status.StartTime,
		"duration":       status.Duration.String(),
		"error_message":  status.ErrorMessage,
	}
}

// ForceSetup forces infrastructure setup (admin function)
func (s *Service) ForceSetup() error {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use pointer, no copying
	s.logger.Info("Forcing infrastructure setup")
	return w.ForceSetup()
}

// WaitForCompletion waits for infrastructure setup to complete (with timeout)
func (s *Service) WaitForCompletion(timeoutSeconds int) error {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use pointer, no copying
	s.logger.Infof("Waiting for infrastructure setup completion (timeout: %ds)", timeoutSeconds)

	// This is a simple polling approach - in production you might want to use channels/events
	for i := 0; i < timeoutSeconds; i++ {
		if completed, err := s.IsSetupCompleted(); err != nil {
			return fmt.Errorf("error checking completion status: %w", err)
		} else if completed {
			s.logger.Info("Infrastructure setup completed")
			return nil
		}

		// Check every second
		select {
		case <-w.Worker.StopChan:
			return fmt.Errorf("worker stopped before completion")
		default:
			// Continue polling
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for infrastructure setup completion")
}

// ScheduleDelete schedules infrastructure deletion to be processed by cron job
func (s *Service) ScheduleDelete() error {
	// Use the models.Worker directly without copying it
	worker := s.worker
	w := &Worker{Worker: worker} // Use pointer, no copying
	s.logger.Warn("Scheduling infrastructure deletion")
	return w.ScheduleDelete()
}
