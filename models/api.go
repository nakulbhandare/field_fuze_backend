package models

// APIResponse is a generic structure for all API responses
type APIResponse struct {
	Status  string      `json:"status"`            // "success" or "error"
	Code    int         `json:"code"`              // HTTP status code (200, 400, 500, etc.)
	Message string      `json:"message,omitempty"` // Human-readable message
	Data    interface{} `json:"data,omitempty"`    // Any response data (can be map, struct, list, etc.)
	Error   *APIError   `json:"error,omitempty"`   // Detailed error info (nil if success)
}

// APIError holds detailed error information
type APIError struct {
	Type    string `json:"type,omitempty"`    // e.g., "ValidationError", "DatabaseError"
	Details string `json:"details,omitempty"` // More context about the error
	Field   string `json:"field,omitempty"`   // For validation errors (which field failed)
}
