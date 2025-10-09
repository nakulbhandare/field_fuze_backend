package models

import "time"

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusInactive  OrganizationStatus = "inactive"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
)

// Organization represents an organization in the system
type Organization struct {
	ID          string             `json:"id" dynamodbav:"id" validate:"omitempty,uuid4"`
	Name        string             `json:"name" dynamodbav:"name" validate:"required,min=2,max=100"`
	Description string             `json:"description,omitempty" dynamodbav:"description,omitempty" validate:"omitempty,max=500"`
	Status      OrganizationStatus `json:"status" dynamodbav:"status" validate:"required,oneof=active inactive suspended"`
	CreatedAt   time.Time          `json:"created_at" dynamodbav:"created_at" validate:"omitempty"`
	UpdatedAt   time.Time          `json:"updated_at" dynamodbav:"updated_at" validate:"omitempty"`
	CreatedBy   string             `json:"created_by" dynamodbav:"created_by" validate:"omitempty"`                     // Audit fields
	UpdatedBy   string             `json:"updated_by,omitempty" dynamodbav:"updated_by,omitempty" validate:"omitempty"` // Contact information
	Email       string             `json:"email,omitempty" dynamodbav:"email,omitempty" validate:"omitempty,email"`
	Phone       string             `json:"phone,omitempty" dynamodbav:"phone,omitempty" validate:"omitempty,e164"`
	Address     string             `json:"address,omitempty" dynamodbav:"address,omitempty" validate:"omitempty,max=200"`
	City        string             `json:"city,omitempty" dynamodbav:"city,omitempty" validate:"omitempty,min=2,max=50"`
	State       string             `json:"state,omitempty" dynamodbav:"state,omitempty" validate:"omitempty,min=2,max=50"`
	Country     string             `json:"country,omitempty" dynamodbav:"country,omitempty" validate:"omitempty,min=2,max=50"`
	PostalCode  string             `json:"postal_code,omitempty" dynamodbav:"postal_code,omitempty" validate:"omitempty,min=3,max=20"` // Business details
	Industry    string             `json:"industry,omitempty" dynamodbav:"industry,omitempty" validate:"omitempty,min=2,max=50"`
}
