package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// RoleAssignment represents a role assignment in JWT
type RoleAssignment struct {
	RoleID      string            `json:"role_id"`
	RoleName    string            `json:"role_name"`
	Level       int               `json:"level"`
	Permissions []string          `json:"permissions"`
	Context     map[string]string `json:"context"`
	AssignedAt  time.Time         `json:"assigned_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// UserContext represents user context in JWT
type UserContext struct {
	OrganizationID string `json:"organization_id,omitempty"`
	CustomerID     string `json:"customer_id,omitempty"`
	WorkerID       string `json:"worker_id,omitempty"`
}

// Role represents a role from roles.json
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	Permissions []string `json:"permissions"`
}

// RolesConfig represents the structure of roles.json
type RolesConfig struct {
	Roles map[string]Role `json:"roles"`
}

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID   string     `json:"user_id"`
	Email    string     `json:"email"`
	Username string     `json:"username"`
	Role     UserRole   `json:"role"` // Keep for backward compatibility
	Status   UserStatus `json:"status"`

	// Enhanced role information
	Roles   []RoleAssignment `json:"roles"`
	Context UserContext      `json:"context"`

	jwt.RegisteredClaims
}
