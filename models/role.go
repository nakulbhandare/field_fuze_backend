package models

import "time"

// RoleStatus represents the status of a role
type RoleStatus string

const (
	RoleStatusActive   RoleStatus = "active"
	RoleStatusInactive RoleStatus = "inactive"
	RoleStatusArchived RoleStatus = "archived"
)

// Role represents a role in the system
type Role struct {
	ID          string                 `json:"id" dynamodbav:"id"`
	Name        string                 `json:"name" dynamodbav:"name"`
	Description string                 `json:"description" dynamodbav:"description"`
	Level       int                    `json:"level" dynamodbav:"level"`
	Permissions []string               `json:"permissions" dynamodbav:"permissions"`
	Status      RoleStatus             `json:"status" dynamodbav:"status"`
	CreatedAt   time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" dynamodbav:"updated_at"`
	CreatedBy   string                 `json:"created_by" dynamodbav:"created_by"`
	UpdatedBy   string                 `json:"updated_by" dynamodbav:"updated_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// CreateRoleRequest represents the request structure for creating a role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required" example:"Admin"`
	Description string   `json:"description" binding:"required" example:"Administrator role with full access"`
	Level       int      `json:"level" binding:"required,min=1,max=10" example:"5"`
	Permissions []string `json:"permissions" binding:"required" example:"[\"read\", \"write\", \"delete\"]"`
}

// UpdateRoleRequest represents the request structure for updating a role
type UpdateRoleRequest struct {
	Name        string     `json:"name,omitempty" example:"Updated Admin"`
	Description string     `json:"description,omitempty" example:"Updated administrator role"`
	Level       *int       `json:"level,omitempty" example:"6"`
	Permissions []string   `json:"permissions,omitempty" example:"[\"read\", \"write\", \"delete\", \"admin\"]"`
	Status      RoleStatus `json:"status,omitempty" example:"active"`
}