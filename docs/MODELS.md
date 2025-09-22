# Models

The Models layer defines the data structures and business entities used throughout the application. Models represent the core domain objects and provide the contract between different layers of the application.

## 📋 Table of Contents

- [Overview](#overview)
- [Model Categories](#model-categories)
- [Core Models](#core-models)
- [API Models](#api-models)
- [Worker Models](#worker-models)
- [Authentication Models](#authentication-models)
- [Validation](#validation)
- [Best Practices](#best-practices)

## Overview

Models in FieldFuze serve multiple purposes:
- **Domain Entities**: Core business objects (User, Role, etc.)
- **API Contracts**: Request/response structures for REST endpoints
- **Database Mapping**: DynamoDB attribute mapping with struct tags
- **Validation Rules**: Input validation using struct tags and custom validators

## Model Categories

### 1. Domain Models
Core business entities that represent the main concepts in the system:
- User management (User, Role)
- Infrastructure (Worker, ExecutionResult)
- Configuration (Config, WorkerConfig)

### 2. API Models
Request and response structures for REST endpoints:
- APIResponse, APIError
- Request DTOs (RegisterUser, UpdateRoleRequest)
- Response DTOs (with pagination metadata)

### 3. Authentication Models
Security and authentication related structures:
- JWTClaims, UserContext
- Permission and role management

### 4. Worker Models
Infrastructure automation and background job models:
- Worker execution status and results
- Progress tracking and error handling

## Core Models

### User Model

**File**: `models/user.go:24-46`

```go
type User struct {
    ID                       string                 `json:"id" dynamodbav:"id"`
    CreatedAt                time.Time              `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt                time.Time              `json:"updated_at" dynamodbav:"updated_at"`
    Email                    string                 `json:"email" dynamodbav:"email"`
    EmailVerified            bool                   `json:"email_verified" dynamodbav:"email_verified"`
    FailedLoginAttempts      int                    `json:"failed_login_attempts" dynamodbav:"failed_login_attempts"`
    FirstName                string                 `json:"first_name" dynamodbav:"first_name"`
    LastName                 string                 `json:"last_name" dynamodbav:"last_name"`
    Password                 string                 `json:"password" dynamodbav:"password_hash,secret"`
    Status                   UserStatus             `json:"status" dynamodbav:"status"`
    Username                 string                 `json:"username" dynamodbav:"username"`
    Roles                    []RoleAssignment       `json:"roles" dynamodbav:"roles"`
    Phone                    *string                `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
    Role                     UserRole               `json:"role,omitempty" dynamodbav:"role,omitempty"` // Backward compatibility
    LastLoginAt              *time.Time             `json:"last_login_at,omitempty" dynamodbav:"last_login_at,omitempty"`
    AccountLockedUntil       *time.Time             `json:"account_locked_until,omitempty" dynamodbav:"account_locked_until,omitempty"`
    EmailVerificationToken   *string                `json:"-" dynamodbav:"email_verification_token,omitempty"`
    PasswordResetToken       *string                `json:"-" dynamodbav:"password_reset_token,omitempty"`
    PasswordResetTokenExpiry *time.Time             `json:"-" dynamodbav:"password_reset_token_expiry,omitempty"`
    Preferences              map[string]interface{} `json:"preferences,omitempty" dynamodbav:"preferences,omitempty"`
}
```

#### User Status Constants
```go
type UserStatus string

const (
    UserStatusActive              UserStatus = "active"
    UserStatusInactive            UserStatus = "inactive"
    UserStatusSuspended           UserStatus = "suspended"
    UserStatusPendingVerification UserStatus = "pending_verification"
)
```

#### User Role Constants
```go
type UserRole string

const (
    UserRoleUser      UserRole = "user"
    UserRoleAdmin     UserRole = "admin"
    UserRoleModerator UserRole = "moderator"
)
```

### Role Models

**File**: `models/role.go`

#### RoleAssignment Model
```go
type RoleAssignment struct {
    RoleID      string            `json:"role_id,omitempty" dynamodbav:"role_id" validate:"omitempty,uuid4"`
    RoleName    string            `json:"role_name" dynamodbav:"role_name" validate:"required,min=2,max=50"`
    Level       int               `json:"level" dynamodbav:"level" validate:"required,min=1,max=10"`
    Permissions []string          `json:"permissions" dynamodbav:"permissions" validate:"required,min=1,dive,oneof=read write delete admin manage create update view"`
    Context     map[string]string `json:"context,omitempty" dynamodbav:"context,omitempty"`
    AssignedAt  time.Time         `json:"assigned_at,omitempty" dynamodbav:"assigned_at" validate:"omitempty"`
    ExpiresAt   *time.Time        `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty" validate:"omitempty"`
}
```

#### Role Model
```go
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
```

#### Role Status Constants
```go
type RoleStatus string

const (
    RoleStatusActive   RoleStatus = "active"
    RoleStatusInactive RoleStatus = "inactive"
    RoleStatusArchived RoleStatus = "archived"
)
```

### Configuration Model

**File**: `models/config.go:5-50`

```go
type Config struct {
    // Application
    AppName    string `mapstructure:"app_name"`
    AppVersion string `mapstructure:"app_version"`
    AppEnv     string `mapstructure:"app_env"`
    AppHost    string `mapstructure:"app_host"`
    AppPort    string `mapstructure:"app_port"`
    
    // JWT
    JWTSecret    string        `mapstructure:"jwt_secret"`
    JWTExpiresIn time.Duration `mapstructure:"jwt_expires_in"`
    
    // Security & Permission Settings
    GracefulPermissionDegradation bool `mapstructure:"graceful_permission_degradation"`
    PermissionCacheTTLSeconds     int  `mapstructure:"permission_cache_ttl_seconds"`
    StrictRoleValidation          bool `mapstructure:"strict_role_validation"`
    LogPermissionChanges          bool `mapstructure:"log_permission_changes"`
    
    // AWS
    AWSRegion           string `mapstructure:"aws_region"`
    AWSAccessKeyID      string `mapstructure:"aws_access_key_id"`
    AWSSecretAccessKey  string `mapstructure:"aws_secret_access_key"`
    DynamoDBEndpoint    string `mapstructure:"dynamodb_endpoint"`
    DynamoDBTablePrefix string `mapstructure:"dynamodb_table_prefix"`
    
    // Telnyx
    TelnyxAPIKey    string `mapstructure:"telnyx_api_key"`
    TelnyxAppID     string `mapstructure:"telnyx_app_id"`
    TelnyxPublicKey string `mapstructure:"telnyx_public_key"`
    
    // Logging
    LogLevel  string `mapstructure:"log_level"`
    LogFormat string `mapstructure:"log_format"`
    
    // CORS
    CORSOrigins []string `mapstructure:"cors_origins"`
    
    // Rate Limiting
    RateLimitRequestsPerMinute int `mapstructure:"rate_limit_requests_per_minute"`
    
    // Base Path
    BasePath string `mapstructure:"basePath"`
    
    Tables []string `mapstructure:"tables"`
}
```

## API Models

### Standard API Response

**File**: `models/api.go:3-17`

```go
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
```

### Request DTOs

#### User Registration
```go
// RegisterUser represents the request structure for user registration
type RegisterUser struct {
    Email       string `json:"email" binding:"required,email" example:"user@example.com" description:"User email address"`
    Username    string `json:"username" binding:"required" example:"john_doe" description:"Desired username"`
    Password    string `json:"password" binding:"required,min=8" example:"securePassword123" description:"User password (minimum 8 characters)"`
    FirstName   string `json:"first_name" binding:"required" example:"John" description:"First name"`
    LastName    string `json:"last_name" binding:"required" example:"Doe" description:"Last name"`
    Phone       string `json:"phone,omitempty" example:"+1234567890" description:"Phone number (optional)"`
    CompanyName string `json:"company_name,omitempty" example:"Acme Corp" description:"Company name (optional)"`
}
```

#### Role Update
```go
// UpdateRoleRequest represents the request structure for updating a role
type UpdateRoleRequest struct {
    Name        string     `json:"name,omitempty" example:"Updated Admin"`
    Description string     `json:"description,omitempty" example:"Updated administrator role"`
    Level       *int       `json:"level,omitempty" example:"6"`
    Permissions []string   `json:"permissions,omitempty" example:"[\"read\", \"write\", \"delete\", \"admin\"]"`
    Status      RoleStatus `json:"status,omitempty" example:"active"`
}
```

### Response Patterns

#### Success Response Example
```json
{
    "status": "success",
    "code": 200,
    "message": "User retrieved successfully",
    "data": {
        "id": "user-123",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe"
    }
}
```

#### Error Response Example
```json
{
    "status": "error",
    "code": 400,
    "message": "Validation failed",
    "error": {
        "type": "ValidationError",
        "details": "email is required; password must be at least 8 characters",
        "field": "email"
    }
}
```

#### Paginated Response Example
```json
{
    "status": "success",
    "code": 200,
    "message": "Users retrieved successfully",
    "data": {
        "users": [...],
        "pagination": {
            "page": 1,
            "limit": 10,
            "total": 100,
            "total_pages": 10,
            "has_next": true,
            "has_previous": false
        }
    }
}
```

## Worker Models

### Worker Status

**File**: `models/worker.go:104-136`

```go
type WorkerStatus string

const (
    // Initial states
    StatusIdle                WorkerStatus = "idle"
    StatusInitializing       WorkerStatus = "initializing"
    
    // Legacy state (keeping for backward compatibility)
    StatusRunning           WorkerStatus = "running"
    
    // Setup phases
    StatusCreatingTables     WorkerStatus = "creating_tables"
    StatusWaitingForTables   WorkerStatus = "waiting_for_tables"
    StatusCreatingIndexes    WorkerStatus = "creating_indexes"
    StatusWaitingForIndexes  WorkerStatus = "waiting_for_indexes"
    
    // Validation phases
    StatusValidating         WorkerStatus = "validating"
    StatusFixingIssues       WorkerStatus = "fixing_issues"
    StatusRevalidating       WorkerStatus = "revalidating"
    
    // Terminal states
    StatusCompleted         WorkerStatus = "completed"
    StatusFailed            WorkerStatus = "failed"
    StatusRetrying          WorkerStatus = "retrying"
    
    // Deletion states
    StatusDeletionScheduled WorkerStatus = "deletion_scheduled"
    StatusDeleting          WorkerStatus = "deleting"
    StatusDeleted           WorkerStatus = "deleted"
    StatusDeletionFailed    WorkerStatus = "deletion_failed"
)
```

### Execution Result

**File**: `models/worker.go:138-167`

```go
type ExecutionResult struct {
    Success        bool                   `json:"success"`
    Status         WorkerStatus           `json:"status"`
    Phase          string                 `json:"phase,omitempty"`          // Current operation phase
    StartTime      time.Time              `json:"start_time"`
    EndTime        *time.Time             `json:"end_time,omitempty"`
    Duration       time.Duration          `json:"duration"`
    
    // Progress tracking
    Progress       *ProgressInfo          `json:"progress,omitempty"`
    
    // Resource tracking
    TablesCreated  []TableStatus          `json:"tables_created"`
    IndexesCreated []IndexStatus          `json:"indexes_created"`
    
    // Error handling
    ErrorMessage   string                 `json:"error_message,omitempty"`
    LastError      *ErrorInfo             `json:"last_error,omitempty"`
    RetryCount     int                    `json:"retry_count"`
    
    // Context
    Environment    string                 `json:"environment"`
    Metadata       map[string]interface{} `json:"metadata"`
    
    // Health indicators
    HealthStatus   string                 `json:"health_status,omitempty"`   // healthy, degraded, unhealthy, provisioning
    NextAction     string                 `json:"next_action,omitempty"`     // What will happen next
    EstimatedTime  *time.Duration         `json:"estimated_time,omitempty"`  // Estimated completion time
}
```

### Progress Tracking

```go
// ProgressInfo tracks execution progress
type ProgressInfo struct {
    CurrentStep    int    `json:"current_step"`
    TotalSteps     int    `json:"total_steps"`
    StepName       string `json:"step_name"`
    Percentage     int    `json:"percentage"`
}

// TableStatus represents enhanced table status information
type TableStatus struct {
    Name           string    `json:"name"`
    Status         string    `json:"status"`         // CREATING, ACTIVE, FAILED
    CreatedAt      time.Time `json:"created_at"`
    BecameActiveAt *time.Time `json:"became_active_at,omitempty"`
    IndexCount     int       `json:"index_count"`
    ExpectedIndexes int      `json:"expected_indexes"`
    Indexes        []IndexStatus `json:"indexes,omitempty"`
    BillingMode    string    `json:"billing_mode,omitempty"`
    Tags           map[string]string `json:"tags,omitempty"`
}

// IndexStatus represents index creation status
type IndexStatus struct {
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Worker Configuration

```go
// WorkerConfig holds configuration for the infrastructure worker
type WorkerConfig struct {
    // Cron schedule
    CronSchedule string `json:"cron_schedule"`
    
    // Lock settings
    LockTimeout       time.Duration `json:"lock_timeout"`
    LockRetryInterval time.Duration `json:"lock_retry_interval"`
    
    // Retry settings
    MaxRetries        int           `json:"max_retries"`
    RetryDelay        time.Duration `json:"retry_delay"`
    BackoffMultiplier float64       `json:"backoff_multiplier"`
    
    // Environment settings
    Environment    string   `json:"environment"`
    RequiredTables []string `json:"required_tables"`
    
    // Paths
    LockFilePath   string `json:"lock_file_path"`
    StatusFilePath string `json:"status_file_path"`
    
    // Feature flags
    DryRun         bool `json:"dry_run"`
    SkipValidation bool `json:"skip_validation"`
    ForceRecreate  bool `json:"force_recreate"`
    RunOnce        bool `json:"run_once"`
    
    // Deletion flags
    DeletionScheduled bool `json:"deletion_scheduled"`
    DeletionRequested bool `json:"deletion_requested"`
}
```

## Authentication Models

### JWT Claims

**File**: `models/auth.go:19-32`

```go
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
```

### User Context

```go
// UserContext represents user context in JWT
type UserContext struct {
    OrganizationID string `json:"organization_id,omitempty"`
    CustomerID     string `json:"customer_id,omitempty"`
    WorkerID       string `json:"worker_id,omitempty"`
}
```

### Infrastructure Request Models

```go
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
```

## Validation

### Struct Tag Validation

The application uses Go Playground Validator for input validation:

```go
type RoleAssignment struct {
    RoleID      string   `validate:"omitempty,uuid4"`
    RoleName    string   `validate:"required,min=2,max=50"`
    Level       int      `validate:"required,min=1,max=10"`
    Permissions []string `validate:"required,min=1,dive,oneof=read write delete admin manage create update view"`
}
```

### Common Validation Tags

- `required` - Field must be present and not empty
- `email` - Must be a valid email address
- `min=X` - Minimum length/value
- `max=X` - Maximum length/value
- `uuid4` - Must be a valid UUID v4
- `oneof=a b c` - Must be one of the specified values
- `dive` - Validates each element in a slice/array

### Gin Binding Tags

For HTTP request binding:

```go
type RegisterUser struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Username string `json:"username" binding:"required"`
}
```

### DynamoDB Mapping Tags

For database persistence:

```go
type User struct {
    ID        string    `json:"id" dynamodbav:"id"`
    Email     string    `json:"email" dynamodbav:"email"`
    Password  string    `json:"password" dynamodbav:"password_hash,secret"`
    CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
}
```

## Best Practices

### 1. Model Design

#### Clear Separation of Concerns
```go
// Domain model - internal representation
type User struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
}

// API request model - external input
type CreateUserRequest struct {
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=8"`
    FirstName string `json:"first_name" binding:"required"`
}

// API response model - external output
type UserResponse struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    FirstName string    `json:"first_name"`
    CreatedAt time.Time `json:"created_at"`
    // Password omitted for security
}
```

#### Use Pointer Types for Optional Fields
```go
type UpdateUserRequest struct {
    FirstName *string `json:"first_name,omitempty"`
    LastName  *string `json:"last_name,omitempty"`
    Phone     *string `json:"phone,omitempty"`
}
```

### 2. JSON Handling

#### Omit Empty Fields
```go
type User struct {
    ID          string     `json:"id"`
    Email       string     `json:"email"`
    Phone       *string    `json:"phone,omitempty"`
    LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}
```

#### Hide Sensitive Fields
```go
type User struct {
    Password string `json:"-"` // Never include in JSON
    Secret   string `json:"-" dynamodbav:"secret"`
}
```

### 3. Constants and Enums

#### Use Type-Safe Constants
```go
type Status string

const (
    StatusActive   Status = "active"
    StatusInactive Status = "inactive"
    StatusArchived Status = "archived"
)

// Validation helper
func (s Status) IsValid() bool {
    switch s {
    case StatusActive, StatusInactive, StatusArchived:
        return true
    default:
        return false
    }
}
```

### 4. Time Handling

#### Use Pointers for Optional Times
```go
type Event struct {
    CreatedAt   time.Time  `json:"created_at"`           // Required
    CompletedAt *time.Time `json:"completed_at,omitempty"` // Optional
}
```

#### Consistent Time Format
```go
// Use time.Time for internal handling
// Use RFC3339 format for JSON serialization
// Let Go handle the conversion automatically
```

### 5. Error Models

#### Structured Error Information
```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Code    string `json:"code"`
}

type APIError struct {
    Type        string             `json:"type"`
    Message     string             `json:"message"`
    Details     string             `json:"details,omitempty"`
    Validations []ValidationError  `json:"validations,omitempty"`
}
```

### 6. Embedding and Composition

#### Common Fields
```go
type BaseModel struct {
    ID        string    `json:"id" dynamodbav:"id"`
    CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

type User struct {
    BaseModel
    Email     string `json:"email" dynamodbav:"email"`
    FirstName string `json:"first_name" dynamodbav:"first_name"`
}
```

### 7. Model Transformations

#### Convert Between Models
```go
func (u *User) ToResponse() *UserResponse {
    return &UserResponse{
        ID:        u.ID,
        Email:     u.Email,
        FirstName: u.FirstName,
        CreatedAt: u.CreatedAt,
        // Exclude sensitive fields like Password
    }
}

func (req *CreateUserRequest) ToUser() *User {
    return &User{
        Email:     req.Email,
        FirstName: req.FirstName,
        CreatedAt: time.Now(),
        Status:    UserStatusActive,
    }
}
```

### 8. Documentation

#### Use Swagger Comments
```go
// User represents a user in the system
// @Description User account with authentication and profile information
type User struct {
    // User ID
    // @example "user-123"
    ID string `json:"id" example:"user-123"`
    
    // User email address
    // @example "user@example.com"
    Email string `json:"email" example:"user@example.com"`
}
```

---

**Related Documentation**: [Controllers](CONTROLLERS.md) | [Services](SERVICES.md) | [Repository](REPOSITORY.md) | [API Reference](API.md)