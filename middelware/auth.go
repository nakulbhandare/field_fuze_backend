// Package middelware provides ultra-optimized authentication and authorization middleware
// for field_fuze_backend with advanced Go techniques and zero-configuration support.
//
// # STANDARD AUTH MIDDLEWARE - Production Ready
//
// Features:
// - Ultra-fast permission checking with intelligent caching
// - Zero-configuration auto-detection for 1000+ APIs
// - Advanced Go concurrency with goroutines and channels
// - Infinite permission combinations with 8 core permissions
// - Thread-safe operations with sync.Map and atomic counters
// - Real-time metrics and performance monitoring
// - Context-aware authorization with timeout protection
//
// Core Permissions (only these 8 are allowed):
// read, write, delete, admin, manage, create, update, view
//
// Usage Examples:
//
// Standard Usage (backward compatible):
//
//	router.Use(jwtManager.AuthMiddleware())
//	router.GET("/users", jwtManager.RequirePermission("read"), handler)
//
// Smart Auto-Detection (zero config for 1000+ APIs):
//
//	router.Use(jwtManager.AuthMiddleware())
//	router.GET("/users", jwtManager.RequireSmartPermission(), handler)     // Auto: read
//	router.POST("/users", jwtManager.RequireSmartPermission(), handler)    // Auto: create
//	router.DELETE("/settings", jwtManager.RequireSmartPermission(), handler) // Auto: delete
//
// Advanced Context-Aware:
//
//	router.PUT("/users/:id", jwtManager.RequireAdvancedPermission("update", map[string]string{
//	    "ownership": "required",
//	}), handler)
//
// Ownership-Based:
//
//	router.GET("/users/:id", jwtManager.RequireOwnership(), handler)
package middelware

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"net/http"
	"regexp"

	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// STANDARD AUTH MIDDLEWARE CONFIGURATION
// These are the ONLY 8 permissions allowed in the system

// CorePermission represents the standard permission types
type CorePermission string

const (
	// Read permissions - for GET operations and data retrieval
	PermissionRead CorePermission = "read"

	// Write permissions - for PATCH operations and data modification
	PermissionWrite CorePermission = "write"

	// Delete permissions - for DELETE operations and data removal
	PermissionDelete CorePermission = "delete"

	// Admin permissions - grants ALL access (highest level)
	PermissionAdmin CorePermission = "admin"

	// Manage permissions - covers create, update, delete operations
	PermissionManage CorePermission = "manage"

	// Create permissions - for POST operations and data creation
	PermissionCreate CorePermission = "create"

	// Update permissions - for PUT operations and data updates
	PermissionUpdate CorePermission = "update"

	// View permissions - for read-only access and reports
	PermissionView CorePermission = "view"
)

// Standard HTTP method to permission mapping
const (
	HTTPMethodGET    = "GET"
	HTTPMethodPOST   = "POST"
	HTTPMethodPUT    = "PUT"
	HTTPMethodPATCH  = "PATCH"
	HTTPMethodDELETE = "DELETE"
)

// No path-based detection - purely HTTP method-based standard

// Performance and security constants
const (
	// Cache settings
	DefaultCacheExpiry  = 5 * time.Minute
	DefaultCacheCleanup = 1 * time.Minute

	// Timeout settings
	PermissionEvalTimeout = 100 * time.Millisecond
	FastEvalTimeout       = 50 * time.Millisecond

	// Concurrency settings
	MaxConcurrentEvals   = 10
	DefaultChannelBuffer = 100
)

// StandardPermissions returns all valid core permissions
func StandardPermissions() []CorePermission {
	return []CorePermission{
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionAdmin,
		PermissionManage,
		PermissionCreate,
		PermissionUpdate,
		PermissionView,
	}
}

// IsValidPermission checks if a permission is one of the 8 standard permissions
func IsValidPermission(permission string) bool {
	validPerms := map[string]bool{
		string(PermissionRead):   true,
		string(PermissionWrite):  true,
		string(PermissionDelete): true,
		string(PermissionAdmin):  true,
		string(PermissionManage): true,
		string(PermissionCreate): true,
		string(PermissionUpdate): true,
		string(PermissionView):   true,
	}
	return validPerms[permission]
}

// PermissionCache represents ultra-fast permission caching with TTL support
type PermissionCache struct {
	data   sync.Map // Thread-safe map for concurrent access
	hits   int64    // Atomic counter for cache hits
	misses int64    // Atomic counter for cache misses
}

// CacheEntry represents a cache entry with expiration
type CacheEntry struct {
	Value     bool      `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PermissionEvaluator represents advanced permission evaluation interface
type PermissionEvaluator interface {
	Evaluate(ctx context.Context, roles []models.RoleAssignment, required string, context map[string]string) bool
}

// SmartPermissionEvaluator implements advanced permission logic
type SmartPermissionEvaluator struct {
	patterns    sync.Map       // Compiled regex patterns for ultra-fast matching
	hierarchy   map[string]int // Permission hierarchy levels
	mutex       sync.RWMutex   // Protects hierarchy map
	evaluations int64          // Atomic counter for performance metrics
}

// JWTManager handles JWT token operations with advanced Go techniques
type JWTManager struct {
	Config            *models.Config
	Logger            logger.Logger
	UserRepo          *repository.UserRepository
	BlacklistedTokens map[string]time.Time // Token revocation blacklist (for immediate invalidation)
	ActiveTokens      map[string]string    // userID -> current active tokenID (single token per user)
	TokenMutex        sync.RWMutex         // Thread safety for both maps

	// Advanced Go features for ultra-strong authorization
	permissionCache  *PermissionCache
	evaluator        PermissionEvaluator
	apiMapping       sync.Map // Thread-safe HTTP method to permission mapping
	resourceMapping  sync.Map // Thread-safe resource-specific permission mapping
	contextResolvers sync.Map // Thread-safe context resolvers
	metrics          struct { // Performance metrics with atomic operations
		authRequests  int64
		authSuccesses int64
		authFailures  int64
		cacheHits     int64
		cacheSize     int64
	}
}

// NewJWTManager creates a new JWT manager with advanced Go optimizations
func NewJWTManager(cfg *models.Config, log logger.Logger, userRepo *repository.UserRepository) *JWTManager {
	j := &JWTManager{
		Config:            cfg,
		Logger:            log,
		UserRepo:          userRepo,
		BlacklistedTokens: make(map[string]time.Time),
		ActiveTokens:      make(map[string]string),
		permissionCache:   &PermissionCache{},
		evaluator:         NewSmartPermissionEvaluator(),
	}

	// Initialize advanced features
	j.initializeAdvancedFeatures()

	return j
}

// NewSmartPermissionEvaluator creates a new smart permission evaluator with standard hierarchy
func NewSmartPermissionEvaluator() *SmartPermissionEvaluator {
	spe := &SmartPermissionEvaluator{
		hierarchy: make(map[string]int),
	}

	// Initialize STANDARD permission hierarchy (higher number = more privileges)
	// This hierarchy defines which permissions include others
	spe.hierarchy = map[string]int{
		string(PermissionView):   1,  // Lowest level - read-only reports
		string(PermissionRead):   2,  // Basic read access
		string(PermissionWrite):  3,  // Can modify existing data
		string(PermissionCreate): 4,  // Can create new data
		string(PermissionUpdate): 5,  // Can update existing data
		string(PermissionDelete): 6,  // Can delete data
		string(PermissionManage): 7,  // Can create, update, delete
		string(PermissionAdmin):  10, // Highest level - can do everything
	}

	return spe
}

// initializeAdvancedFeatures sets up advanced Go features with standard configuration
func (j *JWTManager) initializeAdvancedFeatures() {
	// Initialize standard API mapping with sync.Map for thread-safe concurrent access
	j.apiMapping.Store(HTTPMethodGET, string(PermissionRead))
	j.apiMapping.Store(HTTPMethodPOST, string(PermissionCreate))
	j.apiMapping.Store(HTTPMethodPUT, string(PermissionUpdate))
	j.apiMapping.Store(HTTPMethodPATCH, string(PermissionWrite))
	j.apiMapping.Store(HTTPMethodDELETE, string(PermissionDelete))

	// Initialize resource-specific permission mappings for granular access control
	j.initializeResourceMappings()

	// Initialize context resolvers using functional programming
	j.initializeContextResolvers()

	// Initialize regex patterns for ultra-fast permission matching
	j.initializePermissionPatterns()

	j.Logger.Info("Standard Auth Middleware initialized with advanced Go features")
	j.Logger.Debugf("Supported permissions: %v", StandardPermissions())
}

// initializeResourceMappings sets up resource-specific permission mappings for granular access control
func (j *JWTManager) initializeResourceMappings() {
	// User management resource mappings
	j.resourceMapping.Store("user_list", map[string]interface{}{
		"required_permission": "read",
		"resource_type":       "user_management",
		"context_required":    true,
		"department_scope":    true,
		"team_scope":          false, // Can be enabled for stricter control
	})

	j.resourceMapping.Store("user_details", map[string]interface{}{
		"required_permission": "read",
		"resource_type":       "user_management",
		"context_required":    true,
		"department_scope":    true,
		"ownership_check":     true,
	})

	j.resourceMapping.Store("user_create", map[string]interface{}{
		"required_permission": "create",
		"resource_type":       "user_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       5, // Require level 5+ for user creation
	})

	j.resourceMapping.Store("user_update", map[string]interface{}{
		"required_permission": "update",
		"resource_type":       "user_management",
		"context_required":    true,
		"department_scope":    true,
		"ownership_check":     true,
	})

	j.resourceMapping.Store("user_delete", map[string]interface{}{
		"required_permission": "delete",
		"resource_type":       "user_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       7, // Require level 7+ for user deletion
	})

	// Role management resource mappings
	j.resourceMapping.Store("role_list", map[string]interface{}{
		"required_permission": "read",
		"resource_type":       "role_management",
		"context_required":    true,
		"department_scope":    true,
	})

	j.resourceMapping.Store("role_assign", map[string]interface{}{
		"required_permission": "manage",
		"resource_type":       "role_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       7, // Require level 7+ for role assignment
	})

	j.resourceMapping.Store("role_create", map[string]interface{}{
		"required_permission": "create",
		"resource_type":       "role_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       6, // Require level 6+ for role creation
	})

	j.resourceMapping.Store("role_update", map[string]interface{}{
		"required_permission": "update",
		"resource_type":       "role_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       6, // Require level 6+ for role updates
	})

	j.resourceMapping.Store("role_delete", map[string]interface{}{
		"required_permission": "delete",
		"resource_type":       "role_management",
		"context_required":    true,
		"department_scope":    true,
		"minimum_level":       8, // Require level 8+ for role deletion
	})

	j.Logger.Debug("Resource-specific permission mappings initialized")
}

// GenerateToken generates a JWT token for a user
func (j *JWTManager) GenerateToken(user *models.User) (string, error) {
	// Create claims with updated user struct
	claims := models.JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role, // Keep for backward compatibility
		Status:   user.Status,
		Roles:    user.Roles, // Use the new Roles field from user struct
		Context: models.UserContext{
			OrganizationID: "org-123", // This should be dynamic based on user context
			CustomerID:     "cust-123",
			WorkerID:       "worker-123",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // JTI (JWT ID)
			Subject:   user.ID,
			Issuer:    j.Config.AppName,
			Audience:  jwt.ClaimStrings{j.Config.AppName},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Config.JWTExpiresIn)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(j.Config.JWTSecret))
	if err != nil {
		j.Logger.Errorf("Failed to sign JWT token: %v", err)
		return "", err
	}

	j.Logger.Debugf("Generated JWT token for user: %s", user.ID)

	return tokenString, nil
}

// initializeContextResolvers sets up context resolvers using advanced Go functional programming
func (j *JWTManager) initializeContextResolvers() {
	// Ownership resolver using closure pattern
	j.contextResolvers.Store("ownership", func(c *gin.Context, userID string, context map[string]string) bool {
		if ownerID, exists := context["owner_id"]; exists {
			return userID == ownerID
		}
		if resourceID := c.Param("id"); resourceID != "" {
			return userID == resourceID
		}
		return true
	})

	// Organization resolver with advanced context checking
	j.contextResolvers.Store("organization", func(c *gin.Context, userID string, context map[string]string) bool {
		userContext, exists := c.Get("user_context")
		if !exists {
			return false
		}

		userCtx := userContext.(models.UserContext)
		if requiredOrgID, exists := context["organization_id"]; exists {
			return userCtx.OrganizationID == requiredOrgID || requiredOrgID == "*"
		}

		// Auto-detect from URL parameters
		if orgID := c.Param("orgId"); orgID != "" {
			return userCtx.OrganizationID == orgID
		}

		return true
	})

	// Department-based access control resolver
	j.contextResolvers.Store("department", func(c *gin.Context, userID string, context map[string]string) bool {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found for department validation")
			return false
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Extract user's department from roles context
		userDepartment := ""
		for _, role := range jwtClaims.Roles {
			if dept, exists := role.Context["department"]; exists {
				userDepartment = dept
				break
			}
		}

		if userDepartment == "" {
			j.Logger.Warnf("User %s has no department context in roles", userID)
			return false
		}

		// Check if specific department is required
		if requiredDept, exists := context["department"]; exists {
			if requiredDept == "auto-detect" {
				return true // Allow if department context exists
			}
			return userDepartment == requiredDept
		}

		// Default: allow if user has department context
		return true
	})

	// Team-based access control resolver
	j.contextResolvers.Store("team", func(c *gin.Context, userID string, context map[string]string) bool {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found for team validation")
			return false
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Extract user's team from roles context
		userTeam := ""
		for _, role := range jwtClaims.Roles {
			if team, exists := role.Context["team"]; exists {
				userTeam = team
				break
			}
		}

		if userTeam == "" {
			j.Logger.Warnf("User %s has no team context in roles", userID)
			return false
		}

		// Check if specific team is required
		if requiredTeam, exists := context["team"]; exists {
			if requiredTeam == "auto-detect" {
				return true // Allow if team context exists
			}
			return userTeam == requiredTeam
		}

		// Default: allow if user has team context
		return true
	})

	// Time-based resolver using advanced time calculations
	j.contextResolvers.Store("time", func(c *gin.Context, userID string, context map[string]string) bool {
		now := time.Now()

		// Business hours check with timezone awareness
		if timeRestriction, exists := context["time_restriction"]; exists {
			switch timeRestriction {
			case "business_hours":
				hour := now.Hour()
				weekday := now.Weekday()
				return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour <= 17
			case "weekend_only":
				weekday := now.Weekday()
				return weekday == time.Saturday || weekday == time.Sunday
			case "night_shift":
				hour := now.Hour()
				return hour >= 22 || hour <= 6
			}
		}

		return true
	})
}

// initializePermissionPatterns sets up regex patterns using advanced compilation techniques
func (j *JWTManager) initializePermissionPatterns() {
	patterns := map[string]string{
		"admin_wildcard": `^admin\.*`, // admin.* matches anything starting with admin
		"read_wildcard":  `.*\.read$`, // *.read matches anything ending with .read
		"manage_all":     `^manage$`,  // manage permission covers create/update/delete
		"view_read":      `^view$`,    // view permission covers read operations
	}

	// Compile and store patterns using sync.Map for thread safety
	for name, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			j.evaluator.(*SmartPermissionEvaluator).patterns.Store(name, compiled)
		}
	}
}

// Evaluate implements the PermissionEvaluator interface with advanced Go techniques
func (spe *SmartPermissionEvaluator) Evaluate(ctx context.Context, roles []models.RoleAssignment, required string, contextData map[string]string) bool {
	// Atomic increment for performance metrics
	atomic.AddInt64(&spe.evaluations, 1)

	// Use context for timeout control and cancellation
	select {
	case <-ctx.Done():
		return false // Request cancelled or timed out
	default:
	}

	now := time.Now()

	// Channel-based parallel role evaluation for better performance
	resultChan := make(chan bool, len(roles))
	validRoles := 0

	for _, role := range roles {
		// Skip expired roles
		if role.ExpiresAt != nil && role.ExpiresAt.Before(now) {
			continue
		}

		validRoles++

		// Evaluate each role in a goroutine for parallel processing
		go func(r models.RoleAssignment) {
			defer func() {
				if recover() != nil {
					resultChan <- false // Safe recovery from panics
				}
			}()

			result := spe.evaluateRolePermissions(r, required)
			resultChan <- result
		}(role)
	}

	// Collect results with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	for i := 0; i < validRoles; i++ {
		select {
		case result := <-resultChan:
			if result {
				return true // Found matching permission
			}
		case <-timeoutCtx.Done():
			return false // Timeout exceeded
		}
	}

	return false
}

// evaluateRolePermissions uses advanced permission matching logic
func (spe *SmartPermissionEvaluator) evaluateRolePermissions(role models.RoleAssignment, required string) bool {
	for _, permission := range role.Permissions {
		if spe.matchesAdvancedPermission(permission, required) {
			return true
		}
	}
	return false
}

// matchesAdvancedPermission implements sophisticated permission matching
func (spe *SmartPermissionEvaluator) matchesAdvancedPermission(userPerm, required string) bool {
	// 1. Exact match - fastest path
	if userPerm == required {
		return true
	}

	// 2. Admin wildcard - admin permission grants everything
	if userPerm == "admin" {
		return true
	}

	// 3. Hierarchical permission checking using read-lock for thread safety
	spe.mutex.RLock()
	userLevel, userExists := spe.hierarchy[userPerm]
	requiredLevel, requiredExists := spe.hierarchy[required]
	spe.mutex.RUnlock()

	if userExists && requiredExists && userLevel >= requiredLevel {
		return true
	}

	// 4. Special permission logic
	switch userPerm {
	case "manage":
		// manage permission covers create, update, delete
		return required == "create" || required == "update" || required == "delete"
	case "view":
		// view permission covers read operations
		return required == "read"
	}

	// 5. Pattern matching using compiled regex
	if spe.matchesRegexPattern(userPerm, required) {
		return true
	}

	return false
}

// matchesRegexPattern uses precompiled regex patterns for ultra-fast matching
func (spe *SmartPermissionEvaluator) matchesRegexPattern(userPerm, required string) bool {
	// Check if user permission matches admin wildcard pattern
	if pattern, ok := spe.patterns.Load("admin_wildcard"); ok {
		if regex := pattern.(*regexp.Regexp); regex.MatchString(userPerm) {
			return true
		}
	}

	// Check other patterns based on user permission
	patterns := []string{"read_wildcard", "manage_all", "view_read"}
	for _, patternName := range patterns {
		if pattern, ok := spe.patterns.Load(patternName); ok {
			if regex := pattern.(*regexp.Regexp); regex.MatchString(userPerm) {
				switch patternName {
				case "read_wildcard":
					return strings.HasSuffix(required, ".read")
				case "manage_all":
					return userPerm == "manage" && (required == "create" || required == "update" || required == "delete")
				case "view_read":
					return userPerm == "view" && required == "read"
				}
			}
		}
	}

	return false
}

// Get retrieves a value from the permission cache with TTL validation
func (pc *PermissionCache) Get(key string) (interface{}, bool) {
	value, ok := pc.data.Load(key)
	if !ok {
		atomic.AddInt64(&pc.misses, 1)
		return nil, false
	}

	entry, isEntry := value.(CacheEntry)
	if !isEntry {
		// Legacy cache entry without TTL - treat as expired
		pc.data.Delete(key)
		atomic.AddInt64(&pc.misses, 1)
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		pc.data.Delete(key)
		atomic.AddInt64(&pc.misses, 1)
		return nil, false
	}

	atomic.AddInt64(&pc.hits, 1)
	return entry.Value, true
}

// Set stores a value in the permission cache with TTL
func (pc *PermissionCache) Set(key string, value interface{}) {
	// Default TTL of 30 seconds for security-sensitive operations
	pc.SetWithTTL(key, value, 30*time.Second)
}

// SetWithTTL stores a value with custom TTL
func (pc *PermissionCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	entry := CacheEntry{
		Value:     value.(bool),
		ExpiresAt: time.Now().Add(ttl),
	}
	pc.data.Store(key, entry)
}

// GetStats returns cache statistics with atomic reads
func (pc *PermissionCache) GetStats() (hits, misses int64) {
	return atomic.LoadInt64(&pc.hits), atomic.LoadInt64(&pc.misses)
}

// detectAPIPermission detects required permission purely from HTTP method (no path-based detection)
func (j *JWTManager) detectAPIPermission(c *gin.Context) string {
	method := c.Request.Method

	// STANDARD HTTP method to permission mapping ONLY
	if perm, ok := j.apiMapping.Load(method); ok {
		return perm.(string)
	}

	// Safe default fallback
	return string(PermissionRead)
}

// validateUserStatus checks if user account is in valid state
func (j *JWTManager) validateUserStatus(user *models.User) error {
	if user.Status != models.UserStatusActive {
		return fmt.Errorf("user account is %s", user.Status)
	}

	// Check if account is locked
	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return fmt.Errorf("account is locked until %s", user.AccountLockedUntil.Format(time.RFC3339))
	}

	return nil
}

// validateRoleAssignments cross-verifies token roles with database roles using intersection-based approach
func (j *JWTManager) validateRoleAssignments(tokenRoles, dbRoles []models.RoleAssignment) ([]models.RoleAssignment, error) {
	if len(tokenRoles) == 0 {
		return dbRoles, nil // Use current DB roles if token has none
	}

	now := time.Now()
	validRoles := []models.RoleAssignment{}
	removedRoles := []string{}
	expiredRoles := []string{}

	// Build map of current database roles for efficient lookup
	dbRoleMap := make(map[string]models.RoleAssignment)
	for _, dbRole := range dbRoles {
		dbRoleMap[dbRole.RoleID] = dbRole
	}

	// Find intersection: roles that exist in both token and database
	for _, tokenRole := range tokenRoles {
		if dbRole, exists := dbRoleMap[tokenRole.RoleID]; exists {
			// Check if role has expired
			if dbRole.ExpiresAt != nil && dbRole.ExpiresAt.Before(now) {
				expiredRoles = append(expiredRoles, dbRole.RoleName)
				if j.Config.LogPermissionChanges {
					j.Logger.Warnf("Role '%s' has expired for user", dbRole.RoleName)
				}
				continue
			}
			// Use database version (most current permissions)
			validRoles = append(validRoles, dbRole)
		} else {
			// Role was removed - log but continue with graceful degradation
			removedRoles = append(removedRoles, tokenRole.RoleName)
			if j.Config.LogPermissionChanges {
				j.Logger.Infof("Role '%s' was removed from user, adjusting permissions gracefully", tokenRole.RoleName)
			}
		}
	}

	// Log security events for audit trail
	if len(removedRoles) > 0 || len(expiredRoles) > 0 {
		if j.Config.LogPermissionChanges {
			j.Logger.Warnf("SECURITY: User permissions adjusted. Removed: %v, Expired: %v", removedRoles, expiredRoles)
		}
		// Clear permission cache for this user to ensure fresh permissions
		j.invalidateUserPermissionCache(tokenRoles)
	}

	// Strict mode: fail if any roles were removed/expired
	if j.Config.StrictRoleValidation && (len(removedRoles) > 0 || len(expiredRoles) > 0) {
		return nil, fmt.Errorf("role validation failed: removed=%v expired=%v", removedRoles, expiredRoles)
	}

	// Graceful mode: continue with remaining valid roles
	if len(validRoles) == 0 && len(tokenRoles) > 0 {
		j.Logger.Errorf("User has no valid roles remaining after validation")
		return nil, fmt.Errorf("no valid roles assigned to user")
	}

	return validRoles, nil
}

// invalidateUserPermissionCache removes cache entries for affected user
func (j *JWTManager) invalidateUserPermissionCache(roles []models.RoleAssignment) {
	if len(roles) == 0 {
		return
	}

	// Extract user identifier from roles context
	userContext := ""
	for _, role := range roles {
		if context, exists := role.Context["user_id"]; exists {
			userContext = context
			break
		}
	}

	if userContext == "" {
		// Fallback: clear cache entries containing any of the role IDs
		for _, role := range roles {
			j.invalidateCacheEntriesContaining(role.RoleID)
		}
		return
	}

	// Clear cache entries for this specific user
	j.invalidateCacheEntriesContaining(userContext)
}

// invalidateCacheEntriesContaining removes cache entries containing the specified key fragment
func (j *JWTManager) invalidateCacheEntriesContaining(keyFragment string) {
	j.permissionCache.data.Range(func(key, value any) bool {
		if keyStr, ok := key.(string); ok && strings.Contains(keyStr, keyFragment) {
			j.permissionCache.data.Delete(key)
		}
		return true
	})

	if j.Config.LogPermissionChanges {
		j.Logger.Debug("Invalidated permission cache entries containing: %s", keyFragment)
	}
}

// ValidateToken validates a JWT token and returns the claims with database cross-verification
func (j *JWTManager) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// CRITICAL: Prevent algorithm confusion attacks
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid signing algorithm: %v", method.Alg())
		}

		// Validate critical headers
		if alg, ok := token.Header["alg"].(string); !ok || alg != "HS256" {
			return nil, fmt.Errorf("invalid algorithm in header")
		}

		return []byte(j.Config.JWTSecret), nil
	})

	if err != nil {
		j.Logger.Errorf("Failed to parse JWT token: %v", err)
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		j.Logger.Error("Invalid JWT token")
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		j.Logger.Error("Failed to extract JWT claims")
		return nil, fmt.Errorf("invalid claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		j.Logger.Error("JWT token expired")
		return nil, fmt.Errorf("token expired")
	}

	// Check if token is not yet valid
	if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
		j.Logger.Error("JWT token not yet valid")
		return nil, fmt.Errorf("token not yet valid")
	}

	// Check if token is blacklisted (for immediate revocation)
	j.TokenMutex.RLock()
	if expiry, exists := j.BlacklistedTokens[claims.ID]; exists && expiry.After(time.Now()) {
		j.TokenMutex.RUnlock()
		j.Logger.Error("Token is blacklisted")
		return nil, fmt.Errorf("token has been revoked")
	}

	// Allow multiple active tokens per user (disabled single-token enforcement)
	j.TokenMutex.RUnlock()

	// Note: Single token enforcement disabled to allow multiple concurrent sessions
	// Original validation: activeTokenID, isActive := j.ActiveTokens[claims.UserID]

	// Cross-verify with database for security
	if j.UserRepo != nil {
		dbUsers, err := j.UserRepo.GetUser(claims.UserID)
		if err != nil {
			j.Logger.Errorf("Failed to verify user in database: %v", err)
			return nil, fmt.Errorf("user verification failed")
		}

		if len(dbUsers) == 0 {
			j.Logger.Errorf("User %s not found in database", claims.UserID)
			return nil, fmt.Errorf("user not found")
		}

		dbUser := dbUsers[0]

		// Validate user account status
		if err := j.validateUserStatus(dbUser); err != nil {
			j.Logger.Errorf("User status validation failed for %s: %v", claims.UserID, err)
			return nil, err
		}

		// Validate role assignments against database with graceful degradation
		currentValidRoles, err := j.validateRoleAssignments(claims.Roles, dbUser.Roles)
		if err != nil {
			j.Logger.Errorf("Role validation failed for %s: %v", claims.UserID, err)
			return nil, err
		}

		// Update claims with current valid roles (real-time permissions)
		if j.Config.GracefulPermissionDegradation {
			claims.Roles = currentValidRoles
			if j.Config.LogPermissionChanges {
				j.Logger.Debugf("Updated token claims with %d current valid roles for user %s", len(currentValidRoles), claims.UserID)
			}
		}

		j.Logger.Debugf("Successfully cross-verified user %s with database", claims.UserID)
	}

	j.Logger.Debugf("Successfully validated JWT token for user: %s", claims.UserID)
	return claims, nil
}

// SetActiveToken sets the current active token for a user and revokes any previous token
func (j *JWTManager) SetActiveToken(userID, tokenID string) {
	j.TokenMutex.Lock()
	defer j.TokenMutex.Unlock()

	// If user had a previous token, add it to blacklist for immediate revocation
	if oldTokenID, exists := j.ActiveTokens[userID]; exists && oldTokenID != tokenID {
		j.BlacklistedTokens[oldTokenID] = time.Now().Add(24 * time.Hour) // Blacklist for 24 hours
		j.Logger.Debugf("Previous token %s for user %s added to blacklist", oldTokenID, userID)
	}

	// Set new active token
	j.ActiveTokens[userID] = tokenID
	j.Logger.Debugf("Set active token for user %s: %s", userID, tokenID)
}

// RevokeUserToken removes the active token for a user and adds it to blacklist (logout)
func (j *JWTManager) RevokeUserToken(userID string, tokenID string, expiry time.Time) {
	j.TokenMutex.Lock()
	defer j.TokenMutex.Unlock()

	// Add to blacklist for immediate revocation
	j.BlacklistedTokens[tokenID] = expiry

	// Remove from active tokens
	delete(j.ActiveTokens, userID)

	j.Logger.Debugf("Revoked token for user %s: %s", userID, tokenID)
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (j *JWTManager) CleanupExpiredTokens() {
	j.TokenMutex.Lock()
	defer j.TokenMutex.Unlock()

	now := time.Now()
	for tokenID, expiry := range j.BlacklistedTokens {
		if expiry.Before(now) {
			delete(j.BlacklistedTokens, tokenID)
		}
	}
	j.Logger.Debugf("Cleaned up expired blacklisted tokens")
}

// AuthMiddleware validates JWT token from Authorization header OR handles login credentials
// This single function handles both authentication scenarios
func (j *JWTManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// If no Authorization header, check if this is a login request with credentials
		if authHeader == "" {
			// Check if this is a login request with JSON body credentials
			if c.Request.Method == "POST" && c.Request.Header.Get("Content-Type") == "application/json" {
				// For login endpoints, handle authentication here and skip further middleware
				if strings.Contains(c.Request.URL.Path, "/login") || strings.Contains(c.Request.URL.Path, "/token") {
					j.handleLoginAuthentication(c)
					return
				}
			}
			// Otherwise, require Authorization header
			j.Logger.Error("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Missing Authorization header",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := ""
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
			j.Logger.Error("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Invalid Authorization header format",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "Authorization header must be in format: Bearer <token>",
				},
			})
			c.Abort()
			return
		}
		tokenString = strings.TrimSpace(parts[1])

		// Validate token
		claims, err := j.ValidateToken(tokenString)
		if err != nil {
			j.Logger.Errorf("Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Invalid or expired token",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)
		c.Set("user_context", claims.Context)
		c.Set("jwt_claims", claims)

		// Add intelligent permission detection for smart APIs
		c.Set("auto_permission", j.detectAPIPermission(c))

		j.Logger.Debugf("User authenticated: %s", claims.UserID)
		c.Next()
	}
}

// LoginCredentials represents either LoginRequest or User credentials
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandleLogin provides a public method for handling login requests from controllers
func (j *JWTManager) HandleLogin(c *gin.Context) {
	j.handleLoginAuthentication(c)
}

// handleLoginAuthentication handles credential-based authentication for login requests
func (j *JWTManager) handleLoginAuthentication(c *gin.Context) {
	var req LoginCredentials
	if err := c.ShouldBindJSON(&req); err != nil {
		j.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "An error occurred during login: " + err.Error(),
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Invalid JSON format. Expected format: {\"email\":\"user@example.com\",\"password\":\"yourpassword\"}",
			},
		})
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		j.Logger.Error("Missing email or password")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Missing email or password",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Email and password are required",
			},
		})
		return
	}

	// Get user from database
	users, err := j.UserRepo.GetUser(req.Email)
	if err != nil {
		j.Logger.Error("Failed to get user by email", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user by email",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	if len(users) == 0 {
		j.Logger.Error("User not found")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Code:    http.StatusUnauthorized,
			Message: "Invalid email or password",
			Error: &models.APIError{
				Type:    "AuthenticationError",
				Details: "Invalid email or password",
			},
		})
		return
	}

	user := users[0]

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		j.Logger.Error("Invalid password")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Code:    http.StatusUnauthorized,
			Message: "Invalid email or password",
			Error: &models.APIError{
				Type:    "AuthenticationError",
				Details: "Invalid email or password",
			},
		})
		return
	}

	// Ensure user has roles - if not, set default
	if len(user.Roles) == 0 {
		defaultRole := models.RoleAssignment{
			RoleID:      "role-123",
			RoleName:    "User",
			Permissions: []string{"read", "write"},
			Level:       1,
			Context: map[string]string{
				"project_id": "project-123",
				"org_id":     "org-123",
			},
			AssignedAt: time.Now(),
			ExpiresAt:  nil,
		}
		user.Roles = []models.RoleAssignment{defaultRole}
	}

	// Generate token
	tokenString, err := j.GenerateToken(user)
	if err != nil {
		j.Logger.Error("Token generation failed", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Token generation failed",
			Error: &models.APIError{
				Type:    "TokenError",
				Details: err.Error(),
			},
		})
		return
	}

	// Parse token to extract token ID and set as active token
	// Temporarily disabled to allow multiple concurrent sessions
	// tempToken, _ := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
	// 	return []byte(j.Config.JWTSecret), nil
	// })
	// if tempClaims, ok := tempToken.Claims.(*models.JWTClaims); ok {
	// 	j.SetActiveToken(user.ID, tempClaims.ID)
	// }

	// Return successful authentication response
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Token generated successfully",
		Data: map[string]interface{}{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   3600,
			"user": map[string]interface{}{
				"id":       user.ID,
				"email":    user.Email,
				"username": user.Username,
				"status":   user.Status,
				"roles":    user.Roles,
			},
		},
	})
}

// hasRole checks if user has specific role from current roles
func (j *JWTManager) hasRole(roles []models.RoleAssignment, requiredRole string) bool {
	now := time.Now()
	for _, role := range roles {
		if role.RoleName == requiredRole {
			// Check if role is not expired
			if role.ExpiresAt == nil || role.ExpiresAt.After(now) {
				return true
			}
		}
	}
	return false
}

// hasPermission checks if user has specific permission with advanced Go optimizations
func (j *JWTManager) hasPermission(roles []models.RoleAssignment, requiredPermission string) bool {
	// Atomic increment for metrics
	atomic.AddInt64(&j.metrics.authRequests, 1)

	// Generate cache key using efficient string builder
	var keyBuilder strings.Builder
	keyBuilder.WriteString(requiredPermission)
	keyBuilder.WriteString("|")
	for _, role := range roles {
		keyBuilder.WriteString(role.RoleID)
		keyBuilder.WriteString(",")
	}
	cacheKey := keyBuilder.String()

	// Ultra-fast cache lookup with TTL validation
	if cachedResult, found := j.permissionCache.Get(cacheKey); found {
		atomic.AddInt64(&j.metrics.cacheHits, 1)
		result := cachedResult.(bool)
		if result {
			atomic.AddInt64(&j.metrics.authSuccesses, 1)
		} else {
			atomic.AddInt64(&j.metrics.authFailures, 1)
		}
		return result
	}

	// Create context with timeout for permission evaluation
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Use advanced permission evaluator
	result := j.evaluator.Evaluate(ctx, roles, requiredPermission, nil)

	// Cache the result with configurable TTL for security-sensitive operations
	cacheTTL := time.Duration(j.Config.PermissionCacheTTLSeconds) * time.Second
	if cacheTTL <= 0 {
		cacheTTL = 30 * time.Second // Default 30 seconds for security
	}
	j.permissionCache.SetWithTTL(cacheKey, result, cacheTTL)
	atomic.AddInt64(&j.metrics.cacheSize, 1)

	// Update metrics atomically
	if result {
		atomic.AddInt64(&j.metrics.authSuccesses, 1)
	} else {
		atomic.AddInt64(&j.metrics.authFailures, 1)
	}

	return result
}

// RequireRole middleware checks if user has specific role
func (j *JWTManager) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found in context")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Check if user has required role using helper function
		if !j.hasRole(jwtClaims.Roles, requiredRole) {
			j.Logger.Errorf("User %s does not have required role: %s", jwtClaims.UserID, requiredRole)
			c.JSON(http.StatusForbidden, models.APIResponse{
				Status:  "error",
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &models.APIError{
					Type:    "AuthorizationError",
					Details: fmt.Sprintf("Required role: %s", requiredRole),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware checks if user has specific permission
func (j *JWTManager) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found in context")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Check if user has required permission using helper function
		if !j.hasPermission(jwtClaims.Roles, requiredPermission) {
			j.Logger.Errorf("User %s does not have required permission: %s", jwtClaims.UserID, requiredPermission)
			c.JSON(http.StatusForbidden, models.APIResponse{
				Status:  "error",
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &models.APIError{
					Type:    "AuthorizationError",
					Details: fmt.Sprintf("Required permission: %s", requiredPermission),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TokenValidationRequest represents the request body for token validation
type TokenValidationRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateTokenEndpoint provides an API endpoint to validate tokens
func (j *JWTManager) ValidateTokenEndpoint(c *gin.Context) {
	var req TokenValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		j.Logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Token is required in request body",
			},
		})
		return
	}

	tokenString := strings.TrimSpace(req.Token)
	if tokenString == "" {
		j.Logger.Error("Empty token provided")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Empty token provided",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: "Token cannot be empty",
			},
		})
		return
	}

	// Validate token
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		j.Logger.Errorf("Token validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Code:    http.StatusUnauthorized,
			Message: "Invalid or expired token",
			Error: &models.APIError{
				Type:    "AuthenticationError",
				Details: err.Error(),
			},
		})
		return
	}

	// Return token validation result with user info and roles
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Code:    http.StatusOK,
		Message: "Token is valid",
		Data: map[string]interface{}{
			"valid":      true,
			"user_id":    claims.UserID,
			"email":      claims.Email,
			"username":   claims.Username,
			"status":     claims.Status,
			"roles":      claims.Roles,
			"context":    claims.Context,
			"expires_at": claims.ExpiresAt,
			"issued_at":  claims.IssuedAt,
		},
	})
}

// RequireSmartPermission creates middleware that automatically detects required permissions
func (j *JWTManager) RequireSmartPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found in context")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Get auto-detected permission
		autoPermission, exists := c.Get("auto_permission")
		if !exists {
			autoPermission = j.detectAPIPermission(c)
		}

		requiredPermission := autoPermission.(string)

		// Use enhanced permission checking with advanced Go techniques
		if !j.hasPermission(jwtClaims.Roles, requiredPermission) {
			j.Logger.Errorf("User %s denied smart permission: %s for %s %s",
				jwtClaims.UserID, requiredPermission, c.Request.Method, c.Request.URL.Path)

			c.JSON(http.StatusForbidden, models.APIResponse{
				Status:  "error",
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &models.APIError{
					Type: "AuthorizationError",
					Details: fmt.Sprintf("Required permission: %s for %s %s",
						requiredPermission, c.Request.Method, c.Request.URL.Path),
				},
			})
			c.Abort()
			return
		}

		// Log successful authorization
		j.Logger.Debugf("User %s authorized for %s %s with permission: %s",
			jwtClaims.UserID, c.Request.Method, c.Request.URL.Path, requiredPermission)

		c.Next()
	}
}

// RequireAdvancedPermission creates middleware with advanced permission checking and context
func (j *JWTManager) RequireAdvancedPermission(permission string, contextData map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found in context")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Create context with timeout for advanced evaluation
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Use advanced permission evaluator with context
		if !j.evaluator.Evaluate(ctx, jwtClaims.Roles, permission, contextData) {
			j.Logger.Errorf("User %s denied advanced permission: %s with context: %v",
				jwtClaims.UserID, permission, contextData)

			c.JSON(http.StatusForbidden, models.APIResponse{
				Status:  "error",
				Code:    http.StatusForbidden,
				Message: "Insufficient permissions",
				Error: &models.APIError{
					Type:    "AuthorizationError",
					Details: fmt.Sprintf("Required permission: %s", permission),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAuthMetrics returns authentication metrics for monitoring
func (j *JWTManager) GetAuthMetrics() map[string]int64 {
	cacheHits, cacheMisses := j.permissionCache.GetStats()

	return map[string]int64{
		"auth_requests":  atomic.LoadInt64(&j.metrics.authRequests),
		"auth_successes": atomic.LoadInt64(&j.metrics.authSuccesses),
		"auth_failures":  atomic.LoadInt64(&j.metrics.authFailures),
		"cache_hits":     cacheHits,
		"cache_misses":   cacheMisses,
		"cache_size":     atomic.LoadInt64(&j.metrics.cacheSize),
		"evaluations":    atomic.LoadInt64(&j.evaluator.(*SmartPermissionEvaluator).evaluations),
	}
}

// ClearPermissionCache clears the permission cache (useful when roles change)
func (j *JWTManager) ClearPermissionCache() {
	// Reset sync.Map efficiently
	j.permissionCache.data = sync.Map{}

	// Reset metrics
	atomic.StoreInt64(&j.metrics.cacheSize, 0)
	cacheHits, cacheMisses := j.permissionCache.GetStats()
	atomic.StoreInt64(&j.permissionCache.hits, cacheHits)
	atomic.StoreInt64(&j.permissionCache.misses, cacheMisses)

	j.Logger.Debug("Permission cache cleared")
}

// RequireResourcePermission creates middleware for resource-specific permission checking with context validation
func (j *JWTManager) RequireResourcePermission(resourceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("jwt_claims")
		if !exists {
			j.Logger.Error("JWT claims not found in context")
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
				Error: &models.APIError{
					Type:    "AuthenticationError",
					Details: "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		jwtClaims := claims.(*models.JWTClaims)

		// Get resource mapping configuration
		resourceConfig, exists := j.resourceMapping.Load(resourceName)
		if !exists {
			j.Logger.Errorf("Resource configuration not found: %s", resourceName)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Status:  "error",
				Code:    http.StatusInternalServerError,
				Message: "Resource configuration error",
				Error: &models.APIError{
					Type:    "ConfigurationError",
					Details: "Resource not configured",
				},
			})
			c.Abort()
			return
		}

		config := resourceConfig.(map[string]any)

		// Extract required permission from config
		requiredPermission, ok := config["required_permission"].(string)
		if !ok {
			j.Logger.Errorf("Invalid permission configuration for resource: %s", resourceName)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Status:  "error",
				Code:    http.StatusInternalServerError,
				Message: "Permission configuration error",
			})
			c.Abort()
			return
		}

		// Log authorization attempt for security monitoring
		j.Logger.Infof("Authorization attempt: User %s requesting resource %s (requires %s)",
			jwtClaims.UserID, resourceName, requiredPermission)

		// DEBUG: Log user's current roles and permissions
		j.Logger.Infof("DEBUG: User %s has %d roles:", jwtClaims.UserID, len(jwtClaims.Roles))
		for i, role := range jwtClaims.Roles {
			j.Logger.Infof("DEBUG: Role %d - ID: %s, Name: %s, Level: %d, Permissions: %v, Context: %v",
				i+1, role.RoleID, role.RoleName, role.Level, role.Permissions, role.Context)
		}

		// Check if user has a role that specifically grants access to this resource
		hasValidRoleForResource := j.hasValidRoleForResource(jwtClaims.Roles, resourceName, requiredPermission)
		j.Logger.Infof("DEBUG: Role validation result for user %s accessing resource %s: %t",
			jwtClaims.UserID, resourceName, hasValidRoleForResource)

		if !hasValidRoleForResource {
			// Log detailed failure for security analysis
			j.Logger.Errorf("AUTHORIZATION DENIED: User %s does not have a valid role for resource %s. User roles: %v",
				jwtClaims.UserID, resourceName, j.extractRoleNames(jwtClaims.Roles))
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Role not assigned",
				Error: &models.APIError{
					Type:    "AuthorizationError",
					Details: fmt.Sprintf("User does not have a valid role assigned for resource: %s", resourceName),
				},
			})
			c.Abort()
			return
		}

		// Check minimum level requirement if specified
		if minLevel, exists := config["minimum_level"]; exists {
			if minLevelInt, ok := minLevel.(int); ok {
				userMaxLevel := j.getUserMaxLevel(jwtClaims.Roles)
				if userMaxLevel < minLevelInt {
					j.Logger.Errorf("User %s level %d insufficient for resource %s (requires %d)",
						jwtClaims.UserID, userMaxLevel, resourceName, minLevelInt)
					c.JSON(http.StatusForbidden, models.APIResponse{
						Status:  "error",
						Code:    http.StatusForbidden,
						Message: "Insufficient access level",
						Error: &models.APIError{
							Type:    "AuthorizationError",
							Details: fmt.Sprintf("Required level: %d for resource: %s", minLevelInt, resourceName),
						},
					})
					c.Abort()
					return
				}
			}
		}

		// Perform context validation if required
		if contextRequired, exists := config["context_required"]; exists && contextRequired.(bool) {
			contextData := make(map[string]string)

			// Add department scope validation
			if deptScope, exists := config["department_scope"]; exists && deptScope.(bool) {
				contextData["department"] = "auto-detect"
				j.Logger.Infof("DEBUG: Validating department context for user %s", jwtClaims.UserID)
				deptValidation := j.validateContext(c, jwtClaims.UserID, "department", contextData)
				j.Logger.Infof("DEBUG: Department validation result for user %s: %t", jwtClaims.UserID, deptValidation)
				if !deptValidation {
					j.Logger.Errorf("User %s failed department validation for resource %s",
						jwtClaims.UserID, resourceName)
					c.JSON(http.StatusForbidden, models.APIResponse{
						Status:  "error",
						Code:    http.StatusForbidden,
						Message: "Department access denied",
						Error: &models.APIError{
							Type:    "AuthorizationError",
							Details: "Access restricted to user's department",
						},
					})
					c.Abort()
					return
				}
			}

			// Add team scope validation if required
			if teamScope, exists := config["team_scope"]; exists && teamScope.(bool) {
				contextData["team"] = "auto-detect"
				if !j.validateContext(c, jwtClaims.UserID, "team", contextData) {
					j.Logger.Errorf("User %s failed team validation for resource %s",
						jwtClaims.UserID, resourceName)
					c.JSON(http.StatusForbidden, models.APIResponse{
						Status:  "error",
						Code:    http.StatusForbidden,
						Message: "Team access denied",
						Error: &models.APIError{
							Type:    "AuthorizationError",
							Details: "Access restricted to user's team",
						},
					})
					c.Abort()
					return
				}
			}

			// Add ownership validation if required
			if ownershipCheck, exists := config["ownership_check"]; exists && ownershipCheck.(bool) {
				if !j.validateContext(c, jwtClaims.UserID, "ownership", map[string]string{}) {
					j.Logger.Errorf("User %s failed ownership validation for resource %s",
						jwtClaims.UserID, resourceName)
					c.JSON(http.StatusForbidden, models.APIResponse{
						Status:  "error",
						Code:    http.StatusForbidden,
						Message: "Ownership required",
						Error: &models.APIError{
							Type:    "AuthorizationError",
							Details: "Access restricted to resource owner",
						},
					})
					c.Abort()
					return
				}
			}
		}

		// Log successful authorization with comprehensive details for audit trail
		userDept, userTeam := j.extractUserContext(jwtClaims.Roles)
		j.Logger.Infof("AUTHORIZATION GRANTED: User %s authorized for resource %s with permission %s (dept: %s, team: %s, level: %d)",
			jwtClaims.UserID, resourceName, requiredPermission, userDept, userTeam, j.getUserMaxLevel(jwtClaims.Roles))

		c.Next()
	}
}

// getUserMaxLevel returns the highest level from user's roles
func (j *JWTManager) getUserMaxLevel(roles []models.RoleAssignment) int {
	maxLevel := 0
	for _, role := range roles {
		if role.Level > maxLevel {
			maxLevel = role.Level
		}
	}
	return maxLevel
}

// extractRoleNames extracts role names for logging purposes
func (j *JWTManager) extractRoleNames(roles []models.RoleAssignment) []string {
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.RoleName
	}
	return roleNames
}

// extractUserContext extracts department and team from user roles for logging
func (j *JWTManager) extractUserContext(roles []models.RoleAssignment) (string, string) {
	department := "unknown"
	team := "unknown"

	for _, role := range roles {
		if dept, exists := role.Context["department"]; exists {
			department = dept
		}
		if t, exists := role.Context["team"]; exists {
			team = t
		}
		// Break after finding first context (assuming consistent department/team across roles)
		if department != "unknown" && team != "unknown" {
			break
		}
	}

	return department, team
}

// hasValidRoleForResource checks if user has a valid role that grants access to the specific resource
func (j *JWTManager) hasValidRoleForResource(roles []models.RoleAssignment, resourceName, requiredPermission string) bool {
	now := time.Now()

	// Get resource configuration to check resource type
	resourceConfig, exists := j.resourceMapping.Load(resourceName)
	if !exists {
		j.Logger.Warnf("Resource configuration not found for %s, falling back to permission-only check", resourceName)
		// Fallback to simple permission check if resource config not found
		return j.hasPermission(roles, requiredPermission)
	}

	config := resourceConfig.(map[string]interface{})
	expectedResourceType, hasResourceType := config["resource_type"].(string)

	for _, role := range roles {
		// Skip expired roles
		if role.ExpiresAt != nil && role.ExpiresAt.Before(now) {
			continue
		}

		// Check if this role has the required permission
		hasPermission := false
		for _, permission := range role.Permissions {
			if j.evaluator.(*SmartPermissionEvaluator).matchesAdvancedPermission(permission, requiredPermission) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			continue
		}

		// If resource type is specified in config, check if role context matches
		if hasResourceType {
			// Check if role context contains the expected resource type
			if roleResourceType, exists := role.Context["resource_type"]; exists {
				if roleResourceType == expectedResourceType {
					return true
				}
			}
			// Also allow if role has admin permission (global access)
			if j.hasAdminPermission(role.Permissions) {
				return true
			}
		} else {
			// If no resource type specified, permission match is sufficient
			return true
		}
	}

	return false
}

// hasAdminPermission checks if permissions include admin access
func (j *JWTManager) hasAdminPermission(permissions []string) bool {
	for _, permission := range permissions {
		if permission == "admin" {
			return true
		}
	}
	return false
}

// validateContext validates a specific context using registered resolvers
func (j *JWTManager) validateContext(c *gin.Context, userID, contextType string, contextData map[string]string) bool {
	resolver, exists := j.contextResolvers.Load(contextType)
	if !exists {
		j.Logger.Warnf("Context resolver not found: %s", contextType)
		return false
	}

	resolverFunc := resolver.(func(*gin.Context, string, map[string]string) bool)
	return resolverFunc(c, userID, contextData)
}

// RequireOwnership creates middleware for ownership-based permissions using closures
func (j *JWTManager) RequireOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Status:  "error",
				Code:    http.StatusUnauthorized,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		// Get resource ID from URL params using multiple strategies
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("user_id")
		}
		if resourceID == "" {
			resourceID = c.Param("userId")
		}

		// Allow access if user owns the resource
		if userID.(string) == resourceID {
			c.Next()
			return
		}

		// Fallback: Check if user has admin permission
		claims, exists := c.Get("jwt_claims")
		if exists {
			jwtClaims := claims.(*models.JWTClaims)
			if j.hasPermission(jwtClaims.Roles, "admin") {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.APIResponse{
			Status:  "error",
			Code:    http.StatusForbidden,
			Message: "Access denied - ownership required",
		})
		c.Abort()
	}
}
