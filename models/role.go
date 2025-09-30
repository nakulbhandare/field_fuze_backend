package models

import "time"

// RoleAssignment represents a role assignment in the system
type RoleAssignment struct {
	RoleID      string            `json:"role_id,omitempty" dynamodbav:"role_id" validate:"omitempty,uuid4"`
	RoleName    string            `json:"role_name" dynamodbav:"role_name" validate:"required,min=2,max=50"`
	Level       int               `json:"level" dynamodbav:"level" validate:"required,min=1,max=10"`
	Permissions []string          `json:"permissions" dynamodbav:"permissions" validate:"required,min=1,dive,oneof=read write delete admin manage create update view"`
	Context     map[string]string `json:"context,omitempty" dynamodbav:"context,omitempty"`
	AssignedAt  time.Time         `json:"assigned_at,omitempty" dynamodbav:"assigned_at" validate:"omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty" validate:"omitempty"`
}

// RoleStatus represents the status of a role
type RoleStatus string

const (
	RoleStatusActive   RoleStatus = "active"
	RoleStatusInactive RoleStatus = "inactive"
	RoleStatusArchived RoleStatus = "archived"
)

// Role represents a role in the system (kept for repository compatibility)
type Role struct {
	ID          string                 `json:"id,omitempty" dynamodbav:"id" validate:"omitempty,uuid4"`
	Name        string                 `json:"name" dynamodbav:"name" validate:"required,min=2,max=50"`
	Description string                 `json:"description" dynamodbav:"description" validate:"required,min=10,max=500"`
	Level       int                    `json:"level" dynamodbav:"level" validate:"required,min=1,max=10"`
	Permissions []string               `json:"permissions" dynamodbav:"permissions" validate:"required,min=1,dive,oneof=read write delete admin manage create update view"`
	Status      RoleStatus             `json:"status,omitempty" dynamodbav:"status" validate:"omitempty,oneof=active inactive archived"`
	CreatedAt   time.Time              `json:"created_at,omitempty" dynamodbav:"created_at" validate:"omitempty"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty" dynamodbav:"updated_at" validate:"omitempty"`
	CreatedBy   string                 `json:"created_by,omitempty" dynamodbav:"created_by" validate:"omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty" dynamodbav:"updated_by" validate:"omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// UpdateRoleRequest represents the request structure for updating a role
type UpdateRoleRequest struct {
	Name        string     `json:"name,omitempty" example:"Updated Admin"`
	Description string     `json:"description,omitempty" example:"Updated administrator role"`
	Level       *int       `json:"level,omitempty" example:"6"`
	Permissions []string   `json:"permissions,omitempty" example:"[\"read\", \"write\", \"delete\", \"admin\"]"`
	Status      RoleStatus `json:"status,omitempty" example:"active"`
}
