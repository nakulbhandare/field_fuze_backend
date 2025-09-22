package worker

import (
	"encoding/json"
	"fieldfuze-backend/models"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StatusManager embeds models.StatusManager to allow method definitions
type StatusManager struct {
	models.StatusManager
}

// NewStatusManager creates a new status manager
func NewStatusManager(statusPath string) *StatusManager {
	return &StatusManager{
		StatusManager: models.StatusManager{
			StatusFilePath: statusPath,
		},
	}
}

// ToModelsStatusManager returns the embedded models.StatusManager
func (sm *StatusManager) ToModelsStatusManager() *models.StatusManager {
	return &sm.StatusManager
}

func (sm *StatusManager) SaveStatus(result *models.ExecutionResult) error {
	// Ensure status directory exists
	if err := os.MkdirAll(filepath.Dir(sm.StatusFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Update end time if not set
	if result.EndTime == nil && (result.Status == models.StatusCompleted || result.Status == models.StatusFailed) {
		now := time.Now()
		result.EndTime = &now
		result.Duration = now.Sub(result.StartTime)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Write atomically
	tempFile := sm.StatusFilePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp status file: %w", err)
	}

	if err := os.Rename(tempFile, sm.StatusFilePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename status file: %w", err)
	}

	return nil

}

func (sm *StatusManager) LoadStatus() (*models.ExecutionResult, error) {
	data, err := os.ReadFile(sm.StatusFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file: %w", err)
	}

	var result models.ExecutionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	return &result, nil
}

// IsSetupCompleted checks if infrastructure setup is completed
func (sm *StatusManager) IsSetupCompleted() (bool, error) {
	status, err := sm.LoadStatus()
	if err != nil {
		return false, err
	}

	return status.Status == models.StatusCompleted && status.Success, nil
}

// GetLastExecutionTime returns the last execution time
func (sm *StatusManager) GetLastExecutionTime() (time.Time, error) {
	status, err := sm.LoadStatus()
	if err != nil {
		return time.Time{}, err
	}

	return status.StartTime, nil
}

func (sm *StatusManager) UpdateProgress(status models.WorkerStatus, message string, metadata map[string]any) error {
	currentStatus, err := sm.LoadStatus()
	if err != nil {
		// Create new status if loading fails
		currentStatus = &models.ExecutionResult{
			StartTime:      time.Now(),
			TablesCreated:  make([]models.TableStatus, 0),
			IndexesCreated: make([]models.IndexStatus, 0),
			Metadata:       make(map[string]any),
		}
	}

	currentStatus.Status = status
	if message != "" {
		if currentStatus.Metadata == nil {
			currentStatus.Metadata = make(map[string]any)
		}
		currentStatus.Metadata["last_message"] = message
		currentStatus.Metadata["last_update"] = time.Now()
	}

	// Merge metadata
	if metadata != nil {
		if currentStatus.Metadata == nil {
			currentStatus.Metadata = make(map[string]any)
		}
		for k, v := range metadata {
			currentStatus.Metadata[k] = v
		}
	}

	return sm.SaveStatus(currentStatus)
}

// AddTableCreated adds a table to the created list
func (sm *StatusManager) AddTableCreated(tableName string) error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	// Check if table already in list
	for _, table := range status.TablesCreated {
		if table.Name == tableName {
			return nil // Already recorded
		}
	}

	tableStatus := models.TableStatus{
		Name:      tableName,
		Status:    "CREATING",
		CreatedAt: time.Now(),
	}
	status.TablesCreated = append(status.TablesCreated, tableStatus)
	return sm.SaveStatus(status)
}

// AddIndexCreated adds an index to the created list
func (sm *StatusManager) AddIndexCreated(indexName string) error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	// Check if index already in list
	for _, index := range status.IndexesCreated {
		if index.Name == indexName {
			return nil // Already recorded
		}
	}

	indexStatus := models.IndexStatus{
		Name:      indexName,
		Status:    "CREATING",
		CreatedAt: time.Now(),
	}
	status.IndexesCreated = append(status.IndexesCreated, indexStatus)
	return sm.SaveStatus(status)
}

// MarkCompleted marks the setup as completed
func (sm *StatusManager) MarkCompleted() error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	status.Success = true
	status.Status = models.StatusCompleted
	now := time.Now()
	status.EndTime = &now
	status.Duration = now.Sub(status.StartTime)

	return sm.SaveStatus(status)
}

// MarkFailed marks the setup as failed
func (sm *StatusManager) MarkFailed(errorMsg string) error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	status.Success = false
	status.Status = models.StatusFailed
	status.ErrorMessage = errorMsg
	now := time.Now()
	status.EndTime = &now
	status.Duration = now.Sub(status.StartTime)

	return sm.SaveStatus(status)
}

// IncrementRetryCount increments the retry counter
func (sm *StatusManager) IncrementRetryCount() error {
	status, err := sm.LoadStatus()
	if err != nil {
		return err
	}

	status.RetryCount++
	status.Status = models.StatusRetrying

	return sm.SaveStatus(status)
}

// ResetStatus resets the status (useful for forced re-runs)
func (sm *StatusManager) ResetStatus() error {
	return os.Remove(sm.StatusFilePath)
}
