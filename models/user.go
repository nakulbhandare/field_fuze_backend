package models

import "time"

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser      UserRole = "user"
	UserRoleAdmin     UserRole = "admin"
	UserRoleModerator UserRole = "moderator"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusInactive            UserStatus = "inactive"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

// User represents a user in the system
type User struct {
	ID                       string                 `json:"id" dynamodbav:"id"`
	Email                    string                 `json:"email" dynamodbav:"email"`
	Username                 string                 `json:"username" dynamodbav:"username"`
	PasswordHash             string                 `json:"-" dynamodbav:"password_hash"`
	FirstName                string                 `json:"first_name" dynamodbav:"first_name"`
	LastName                 string                 `json:"last_name" dynamodbav:"last_name"`
	Phone                    *string                `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	Role                     UserRole               `json:"role" dynamodbav:"role"`
	Status                   UserStatus             `json:"status" dynamodbav:"status"`
	CreatedAt                time.Time              `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt                time.Time              `json:"updated_at" dynamodbav:"updated_at"`
	LastLoginAt              *time.Time             `json:"last_login_at,omitempty" dynamodbav:"last_login_at,omitempty"`
	FailedLoginAttempts      int                    `json:"failed_login_attempts" dynamodbav:"failed_login_attempts"`
	AccountLockedUntil       *time.Time             `json:"account_locked_until,omitempty" dynamodbav:"account_locked_until,omitempty"`
	EmailVerified            bool                   `json:"email_verified" dynamodbav:"email_verified"`
	EmailVerificationToken   *string                `json:"-" dynamodbav:"email_verification_token,omitempty"`
	PasswordResetToken       *string                `json:"-" dynamodbav:"password_reset_token,omitempty"`
	PasswordResetTokenExpiry *time.Time             `json:"-" dynamodbav:"password_reset_token_expiry,omitempty"`
	Preferences              map[string]interface{} `json:"preferences,omitempty" dynamodbav:"preferences,omitempty"`
}

// RegisterUser represents the request structure for user registration
// @Description User registration request with account details
type RegisterUser struct {
	Email       string `json:"email" binding:"required,email" example:"user@example.com" description:"User email address"`
	Username    string `json:"username" binding:"required" example:"john_doe" description:"Desired username"`
	Password    string `json:"password" binding:"required,min=8" example:"securePassword123" description:"User password (minimum 8 characters)"`
	FirstName   string `json:"first_name" binding:"required" example:"John" description:"First name"`
	LastName    string `json:"last_name" binding:"required" example:"Doe" description:"Last name"`
	Phone       string `json:"phone,omitempty" example:"+1234567890" description:"Phone number (optional)"`
	CompanyName string `json:"company_name,omitempty" example:"Acme Corp" description:"Company name (optional)"`
}