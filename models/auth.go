package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// UserContext represents user context in JWT
type UserContext struct {
	OrganizationID string `json:"organization_id,omitempty"`
	CustomerID     string `json:"customer_id,omitempty"`
	WorkerID       string `json:"worker_id,omitempty"`
}

// RolesConfig represents the structure of roles.json
type RolesConfig struct {
	Roles map[string]RoleAssignment `json:"roles"`
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
