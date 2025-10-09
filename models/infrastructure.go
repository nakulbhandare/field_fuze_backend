package models

import "time"


// ServiceRestartResult represents the result of a service restart operation
type ServiceRestartResult struct {
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"` // in_progress, completed, failed, not_needed
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Output      string    `json:"output,omitempty"` // Command output
	Error       string    `json:"error,omitempty"`  // Error message if failed
}

// WorkerRestartRequest represents the request body for worker restart
type WorkerRestartRequest struct {
	Force bool `json:"force"` // Force restart even if worker is currently running
}
