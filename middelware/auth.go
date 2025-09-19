package middelware

import (
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	Config            *models.Config
	Logger            logger.Logger
	UserRepo          *repository.UserRepository
	BlacklistedTokens map[string]time.Time // Token revocation blacklist (for immediate invalidation)
	ActiveTokens      map[string]string    // userID -> current active tokenID (single token per user)
	TokenMutex        sync.RWMutex         // Thread safety for both maps
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *models.Config, log logger.Logger, userRepo *repository.UserRepository) *JWTManager {
	return &JWTManager{
		Config:            cfg,
		Logger:            log,
		UserRepo:          userRepo,
		BlacklistedTokens: make(map[string]time.Time),
		ActiveTokens:      make(map[string]string),
	}
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

// validateRoleAssignments cross-verifies token roles with database roles
func (j *JWTManager) validateRoleAssignments(tokenRoles, dbRoles []models.RoleAssignment) error {
	if len(tokenRoles) == 0 {
		return nil // No roles to validate
	}

	now := time.Now()
	validRoles := []models.RoleAssignment{}

	// Check each role in token exists in database and is not expired
	for _, tokenRole := range tokenRoles {
		roleFound := false

		for _, dbRole := range dbRoles {
			if tokenRole.RoleID == dbRole.RoleID {
				// Check if role has expired
				if dbRole.ExpiresAt != nil && dbRole.ExpiresAt.Before(now) {
					j.Logger.Errorf("Role '%s' has expired for user", dbRole.RoleName)
					return fmt.Errorf("role '%s' has expired", dbRole.RoleName)
				}
				roleFound = true
				validRoles = append(validRoles, dbRole)
				break
			}
		}

		if !roleFound {
			j.Logger.Errorf("Token validation failed: role '%s' no longer assigned to user", tokenRole.RoleName)
			return fmt.Errorf("role '%s' no longer assigned to user", tokenRole.RoleName)
		}
	}

	// If user has no valid roles remaining, deny access
	if len(validRoles) == 0 && len(tokenRoles) > 0 {
		j.Logger.Errorf("User has no valid roles remaining")
		return fmt.Errorf("no valid roles assigned to user")
	}

	return nil
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

		// Validate role assignments against database
		if err := j.validateRoleAssignments(claims.Roles, dbUser.Roles); err != nil {
			j.Logger.Errorf("Role validation failed for %s: %v", claims.UserID, err)
			return nil, err
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
				fmt.Println("here i'm")
				j.handleLoginAuthentication(c)
				return
			}
			fmt.Println("hererehehehh112121212")
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

		j.Logger.Debugf("User authenticated: %s", claims.UserID)
		c.Next()
	}
}

// LoginCredentials represents either LoginRequest or User credentials
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

// hasPermission checks if user has specific permission from current roles
func (j *JWTManager) hasPermission(roles []models.RoleAssignment, requiredPermission string) bool {
	now := time.Now()
	for _, role := range roles {
		// Skip expired roles
		if role.ExpiresAt != nil && role.ExpiresAt.Before(now) {
			continue
		}

		for _, permission := range role.Permissions {
			if permission == requiredPermission {
				return true
			}
		}
	}
	return false
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
