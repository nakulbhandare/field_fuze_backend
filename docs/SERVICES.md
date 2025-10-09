# Services

The Services layer contains the core business logic of the application. Services orchestrate data access, implement business rules, and provide a clean interface between controllers and repositories.

## 📋 Table of Contents

- [Overview](#overview)
- [Service Architecture](#service-architecture)
- [Available Services](#available-services)
- [Adding New Services](#adding-new-services)
- [Service Patterns](#service-patterns)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Overview

Services in FieldFuze follow the Service Pattern where:
- **Business Logic**: All business rules and domain logic are implemented here
- **Data Orchestration**: Coordinate between multiple repositories when needed
- **Validation**: Business-level validation beyond basic input validation
- **Abstraction**: Provide clean interfaces for controllers to use

## Service Architecture

### Service Factory Pattern
**File**: `services/services.go:1-11`

```go
package services

type Service struct {
    // UserService *UserService
}

func NewService() *Service {
    return &Service{
        // UserService: NewUserService(),
    }
}
```

## Available Services

### 1. UserService

**File**: `services/user_service.go`

Basic user operations service (currently minimal implementation).

#### Structure
```go
type UserService struct {
    ctx  context.Context
    repo *repository.UserRepository
}

func NewUserService(ctx context.Context, repo *repository.UserRepository) *UserService {
    return &UserService{
        ctx:  ctx,
        repo: repo,
    }
}
```

#### Methods
```go
func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
    return s.repo.CreateUser(s.ctx, user)
}
```

### 2. RoleService

**File**: `services/role_service.go`

Comprehensive role management service with validation and business logic.

#### Structure
```go
type RoleService struct {
    roleRepo *repository.RoleRepository
    logger   logger.Logger
}

func NewRoleService(roleRepo *repository.RoleRepository, logger logger.Logger) *RoleService {
    return &RoleService{
        roleRepo: roleRepo,
        logger:   logger,
    }
}
```

#### Key Methods

##### Create Role with Validation
```go
func (s *RoleService) CreateRole(ctx context.Context, roleAssignment *models.RoleAssignment, createdBy string) (*models.RoleAssignment, error) {
    // Business validation
    if err := s.validateCreateRoleAssignment(roleAssignment); err != nil {
        return nil, err
    }
    
    // Normalize data
    roleAssignment.RoleName = strings.TrimSpace(roleAssignment.RoleName)
    
    // Delegate to repository
    return s.roleRepo.CreateRoleAssignment(ctx, roleAssignment)
}
```

##### Role Validation Logic
```go
func (s *RoleService) validateCreateRoleAssignment(roleAssignment *models.RoleAssignment) error {
    if roleAssignment == nil {
        return errors.New("role assignment is required")
    }
    
    if strings.TrimSpace(roleAssignment.RoleName) == "" {
        return errors.New("role name is required")
    }
    
    if len(roleAssignment.RoleName) > 100 {
        return errors.New("role name must be less than 100 characters")
    }
    
    if roleAssignment.Level < 1 || roleAssignment.Level > 10 {
        return errors.New("role level must be between 1 and 10")
    }
    
    if len(roleAssignment.Permissions) == 0 {
        return errors.New("at least one permission is required")
    }
    
    for _, permission := range roleAssignment.Permissions {
        if strings.TrimSpace(permission) == "" {
            return errors.New("permission cannot be empty")
        }
    }
    
    return nil
}
```

##### Get Role by Name
```go
func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
    if strings.TrimSpace(name) == "" {
        return nil, errors.New("role name is required")
    }
    
    roles, err := s.roleRepo.GetRole(name)
    if err != nil {
        return nil, err
    }
    
    if len(roles) == 0 {
        return nil, errors.New("role not found")
    }
    
    return roles[0], nil
}
```

##### Update Role with Business Logic
```go
func (s *RoleService) UpdateRole(id string, req *models.UpdateRoleRequest, updatedBy string) (*models.Role, error) {
    if strings.TrimSpace(id) == "" {
        return nil, errors.New("role ID is required")
    }
    
    if err := s.validateUpdateRoleRequest(req); err != nil {
        return nil, err
    }
    
    // Build update object with only changed fields
    role := &models.Role{
        UpdatedBy: updatedBy,
    }
    
    if req.Name != "" {
        role.Name = strings.TrimSpace(req.Name)
    }
    if req.Description != "" {
        role.Description = strings.TrimSpace(req.Description)
    }
    if req.Level != nil {
        role.Level = *req.Level
    }
    if req.Permissions != nil {
        role.Permissions = req.Permissions
    }
    if req.Status != "" {
        role.Status = req.Status
    }
    
    return s.roleRepo.UpdateRole(id, role)
}
```

##### Status-Based Filtering
```go
func (s *RoleService) GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error) {
    if status == "" {
        return nil, errors.New("status is required")
    }
    
    return s.roleRepo.GetRoleAssignmentsByStatus(status)
}
```

### 3. InfrastructureService

**File**: `services/infrastructure_service.go`

Manages infrastructure automation, worker status, and system health.

#### Structure
```go
type InfrastructureService struct {
    ctx      context.Context
    dbClient *dal.DynamoDBClient
    logger   logger.Logger
    config   *models.Config
}

func NewInfrastructureService(ctx context.Context, dbClient *dal.DynamoDBClient, logger logger.Logger, config *models.Config) *InfrastructureService {
    return &InfrastructureService{
        ctx:      ctx,
        dbClient: dbClient,
        logger:   logger,
        config:   config,
    }
}
```

#### Key Methods

##### Worker Status Management
```go
// getWorkerStatus reads worker status from the status file
func (s *InfrastructureService) getWorkerStatus() (*models.ExecutionResult, error) {
    // Get status file path based on environment
    statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", s.config.AppEnv)
    
    data, err := os.ReadFile(statusFilePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read worker status file: %w", err)
    }
    
    var result models.ExecutionResult
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal worker status: %w", err)
    }
    
    return &result, nil
}

// GetWorkerStatus returns the current worker status with enhanced context (public method for API)
func (s *InfrastructureService) GetWorkerStatus(ctx context.Context) (*models.ExecutionResult, error) {
    s.logger.Debug("Getting detailed worker status")
    
    result, err := s.getWorkerStatus()
    if err != nil {
        return nil, err
    }
    
    // Enrich status with additional context
    s.enrichStatusWithContext(result)
    
    // Update health indicators
    s.updateHealthIndicators(result)
    
    return result, nil
}
```

##### Worker Restart Logic
```go
func (s *InfrastructureService) RestartWorker(ctx context.Context, force bool) (*models.ServiceRestartResult, error) {
    s.logger.Info("Restarting infrastructure worker")
    
    result := &models.ServiceRestartResult{
        ServiceName: "infrastructure-worker",
        StartTime:   time.Now(),
        Status:      "in_progress",
    }
    
    // Check current worker status
    workerStatus, err := s.getWorkerStatus()
    if err != nil {
        s.logger.Warn("Could not get current worker status, proceeding with restart", err)
    }
    
    // If worker is running and force is false, return error
    if !force && workerStatus != nil && workerStatus.Status == "running" {
        result.Status = "failed"
        result.Error = "Worker is currently running. Use force=true to restart anyway"
        result.EndTime = time.Now()
        return result, fmt.Errorf("worker is running")
    }
    
    // Kill existing worker process if running
    if err := s.killWorkerProcess(); err != nil {
        s.logger.Warn("Failed to kill existing worker process", err)
    }
    
    // Reset worker status to allow restart
    if err := s.resetWorkerStatus(); err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        result.EndTime = time.Now()
        return result, err
    }
    
    // Start new worker process
    if err := s.startWorkerProcess(); err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        result.EndTime = time.Now()
        return result, err
    }
    
    result.Status = "completed"
    result.EndTime = time.Now()
    s.logger.Info("Infrastructure worker restarted successfully")
    
    return result, nil
}
```

##### Health Check Implementation
```go
func (s *InfrastructureService) IsWorkerHealthy() (bool, string, error) {
    status, err := s.getWorkerStatus()
    if err != nil {
        return false, "Cannot read worker status", err
    }
    
    // Check if worker is in a healthy state
    switch status.Status {
    case models.StatusCompleted:
        if status.Success {
            return true, "Worker completed successfully", nil
        }
        return false, "Worker completed with errors", nil
    case models.StatusFailed:
        return false, fmt.Sprintf("Worker failed: %s", status.Error), nil
    case models.StatusRunning, models.StatusInitializing:
        return true, "Worker is running", nil
    case models.StatusRetrying:
        if status.RetryCount > 3 {
            return false, "Worker has too many retry attempts", nil
        }
        return true, fmt.Sprintf("Worker is retrying (attempt %d)", status.RetryCount+1), nil
    default:
        return false, fmt.Sprintf("Worker in unknown state: %s", status.Status), nil
    }
}

func (s *InfrastructureService) AutoRestartIfNeeded(ctx context.Context) (*models.AutoRestartResult, error) {
    healthy, reason, err := s.IsWorkerHealthy()
    if err != nil {
        return nil, err
    }
    
    result := &models.AutoRestartResult{
        CheckedAt:    time.Now(),
        WasHealthy:   healthy,
        HealthReason: reason,
        Status:       "not_needed",
    }
    
    if healthy {
        result.Status = "not_needed"
        return result, nil
    }
    
    // Worker is unhealthy, attempt restart
    s.logger.Warn("Worker is unhealthy, attempting auto-restart", "reason", reason)
    
    restartResult, err := s.RestartWorker(ctx, true) // Force restart
    if err != nil {
        result.Status = "failed"
        result.Error = err.Error()
        return result, err
    }
    
    result.Status = "completed"
    result.RestartedAt = &restartResult.EndTime
    
    return result, nil
}
```

## Adding New Services

### Step 1: Create Service File

Create a new service file in the `services/` directory:

```go
// services/your_service.go
package services

import (
    "context"
    "errors"
    "fieldfuze-backend/models"
    "fieldfuze-backend/repository"
    "fieldfuze-backend/utils/logger"
    "strings"
)

type YourService struct {
    ctx    context.Context
    repo   *repository.YourRepository
    logger logger.Logger
}

func NewYourService(ctx context.Context, repo *repository.YourRepository, logger logger.Logger) *YourService {
    return &YourService{
        ctx:    ctx,
        repo:   repo,
        logger: logger,
    }
}

// Business methods
func (s *YourService) CreateYourResource(resource *models.YourResource) (*models.YourResource, error) {
    // Validate business rules
    if err := s.validateYourResource(resource); err != nil {
        return nil, err
    }
    
    // Apply business logic
    resource.Name = strings.TrimSpace(resource.Name)
    
    // Delegate to repository
    return s.repo.Create(s.ctx, resource)
}

func (s *YourService) validateYourResource(resource *models.YourResource) error {
    if resource == nil {
        return errors.New("resource is required")
    }
    
    if strings.TrimSpace(resource.Name) == "" {
        return errors.New("resource name is required")
    }
    
    // Add more business validation rules
    
    return nil
}
```

### Step 2: Add to Service Factory

Update the main service factory if needed:

```go
// services/services.go
type Service struct {
    UserService *UserService
    YourService *YourService  // Add your service
}

func NewService(ctx context.Context, repos *repository.Repositories, logger logger.Logger) *Service {
    return &Service{
        UserService: NewUserService(ctx, repos.User),
        YourService: NewYourService(ctx, repos.Your, logger),  // Initialize
    }
}
```

### Step 3: Integration Points

Update the controller initialization to use your service:

```go
// In controller/controller.go
func NewController(ctx context.Context, cfg *models.Config, log logger.Logger) *Controller {
    // ... existing initialization ...
    
    yourService := services.NewYourService(ctx, yourRepo, log)
    
    return &Controller{
        // ... existing controllers ...
        Your: NewYourController(ctx, yourService, log),
    }
}
```

## Service Patterns

### 1. Validation Pattern

```go
// Centralized validation
func (s *Service) validateEntity(entity *models.Entity) error {
    if entity == nil {
        return errors.New("entity is required")
    }
    
    // Basic field validation
    if strings.TrimSpace(entity.Name) == "" {
        return errors.New("name is required")
    }
    
    // Business rule validation
    if entity.Level < 1 || entity.Level > 10 {
        return errors.New("level must be between 1 and 10")
    }
    
    // Complex business rules
    if entity.Type == "premium" && entity.Level < 5 {
        return errors.New("premium entities must have level 5 or higher")
    }
    
    return nil
}

// Partial validation for updates
func (s *Service) validatePartialEntity(entity *models.PartialEntity) error {
    if entity.Name != nil && strings.TrimSpace(*entity.Name) == "" {
        return errors.New("name cannot be empty")
    }
    
    if entity.Level != nil && (*entity.Level < 1 || *entity.Level > 10) {
        return errors.New("level must be between 1 and 10")
    }
    
    return nil
}
```

### 2. Data Transformation Pattern

```go
func (s *Service) CreateEntity(req *models.CreateEntityRequest, createdBy string) (*models.Entity, error) {
    // Validate request
    if err := s.validateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // Transform request to entity
    entity := &models.Entity{
        Name:        strings.TrimSpace(req.Name),
        Description: strings.TrimSpace(req.Description),
        Status:      models.StatusActive,
        CreatedBy:   createdBy,
        CreatedAt:   time.Now(),
    }
    
    // Apply business logic
    if req.Type == "auto" {
        entity.Type = s.determineAutoType(entity)
    } else {
        entity.Type = req.Type
    }
    
    // Generate computed fields
    entity.Slug = s.generateSlug(entity.Name)
    entity.Priority = s.calculatePriority(entity)
    
    return s.repo.Create(entity)
}

func (s *Service) determineAutoType(entity *models.Entity) string {
    // Business logic to determine type
    if len(entity.Name) > 50 {
        return "detailed"
    }
    return "simple"
}

func (s *Service) generateSlug(name string) string {
    // Convert name to URL-friendly slug
    slug := strings.ToLower(name)
    slug = strings.ReplaceAll(slug, " ", "-")
    // Remove special characters, etc.
    return slug
}
```

### 3. Multi-Repository Orchestration

```go
func (s *Service) CreateUserWithRole(userReq *models.CreateUserRequest, roleID string) (*models.UserWithRole, error) {
    // Start transaction context
    ctx := s.ctx
    
    // Validate role exists
    role, err := s.roleRepo.GetByID(ctx, roleID)
    if err != nil {
        return nil, fmt.Errorf("invalid role: %w", err)
    }
    
    // Create user
    user, err := s.userRepo.Create(ctx, &models.User{
        Email:     userReq.Email,
        FirstName: userReq.FirstName,
        LastName:  userReq.LastName,
        Status:    models.UserStatusActive,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Assign role
    _, err = s.userRepo.AssignRole(ctx, user.ID, roleID)
    if err != nil {
        // Rollback user creation if role assignment fails
        s.userRepo.Delete(ctx, user.ID)
        return nil, fmt.Errorf("failed to assign role: %w", err)
    }
    
    // Return combined result
    return &models.UserWithRole{
        User: user,
        Role: role,
    }, nil
}
```

### 4. Caching Pattern

```go
type ServiceWithCache struct {
    repo   *repository.Repository
    cache  map[string]*models.Entity
    logger logger.Logger
}

func (s *ServiceWithCache) GetByID(id string) (*models.Entity, error) {
    // Check cache first
    if entity, exists := s.cache[id]; exists {
        s.logger.Debug("Cache hit for entity", "id", id)
        return entity, nil
    }
    
    // Cache miss, get from repository
    entity, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache[id] = entity
    s.logger.Debug("Cache updated for entity", "id", id)
    
    return entity, nil
}

func (s *ServiceWithCache) Update(id string, updates *models.EntityUpdates) (*models.Entity, error) {
    entity, err := s.repo.Update(id, updates)
    if err != nil {
        return nil, err
    }
    
    // Invalidate cache
    delete(s.cache, id)
    s.logger.Debug("Cache invalidated for entity", "id", id)
    
    return entity, nil
}
```

### 5. Event-Driven Pattern

```go
type EventService struct {
    repo      *repository.Repository
    publisher EventPublisher
    logger    logger.Logger
}

func (s *EventService) CreateEntity(entity *models.Entity) (*models.Entity, error) {
    // Create entity
    created, err := s.repo.Create(entity)
    if err != nil {
        return nil, err
    }
    
    // Publish creation event
    event := &models.EntityCreatedEvent{
        EntityID:  created.ID,
        EntityType: created.Type,
        CreatedBy: created.CreatedBy,
        CreatedAt: created.CreatedAt,
    }
    
    if err := s.publisher.Publish("entity.created", event); err != nil {
        s.logger.Error("Failed to publish entity created event", err)
        // Don't fail the operation for event publishing errors
    }
    
    return created, nil
}

func (s *EventService) DeleteEntity(id string) error {
    // Get entity before deletion for event
    entity, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    
    // Delete entity
    if err := s.repo.Delete(id); err != nil {
        return err
    }
    
    // Publish deletion event
    event := &models.EntityDeletedEvent{
        EntityID:   id,
        EntityType: entity.Type,
        DeletedAt:  time.Now(),
    }
    
    s.publisher.Publish("entity.deleted", event)
    
    return nil
}
```

## Error Handling

### Service-Level Error Types

```go
var (
    ErrEntityNotFound     = errors.New("entity not found")
    ErrEntityExists       = errors.New("entity already exists")
    ErrInvalidData        = errors.New("invalid data provided")
    ErrBusinessRule       = errors.New("business rule violation")
    ErrInsufficientAccess = errors.New("insufficient access")
)

// Custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

type BusinessRuleError struct {
    Rule    string
    Message string
}

func (e *BusinessRuleError) Error() string {
    return fmt.Sprintf("business rule '%s' violated: %s", e.Rule, e.Message)
}
```

### Error Wrapping Pattern

```go
func (s *Service) ProcessEntity(id string) (*models.Entity, error) {
    entity, err := s.repo.GetByID(id)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return nil, fmt.Errorf("%w: entity with id %s", ErrEntityNotFound, id)
        }
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    
    if err := s.validateBusinessRules(entity); err != nil {
        return nil, fmt.Errorf("business validation failed: %w", err)
    }
    
    // Process entity
    processed, err := s.applyProcessing(entity)
    if err != nil {
        return nil, fmt.Errorf("processing failed for entity %s: %w", id, err)
    }
    
    return processed, nil
}
```

## Best Practices

### 1. Single Responsibility
- Each service should have a clear, single purpose
- Keep services focused on specific business domains
- Split large services into smaller, focused ones

### 2. Dependency Injection
- Accept dependencies through constructor parameters
- Use interfaces for dependencies to enable testing
- Don't create dependencies inside services

### 3. Validation
- Implement comprehensive business validation
- Separate validation from data access
- Return clear, actionable error messages

### 4. Error Handling
- Use custom error types for different scenarios
- Wrap errors with context
- Log errors appropriately without exposing sensitive data

### 5. Testing
- Write unit tests for all business logic
- Mock dependencies for isolated testing
- Test both happy path and error scenarios

### 6. Logging
- Log important business operations
- Include relevant context in log messages
- Use appropriate log levels

### 7. Performance
- Be mindful of repository calls
- Implement caching where appropriate
- Use pagination for large result sets

### 8. Consistency
- Follow consistent patterns across services
- Use similar naming conventions
- Maintain consistent error handling approaches

---

**Related Documentation**: [Controllers](CONTROLLERS.md) | [Repository](REPOSITORY.md) | [Models](MODELS.md) | [Workers](WORKERS.md)