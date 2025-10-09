# Utilities

The Utilities layer provides common helper functions, configuration management, logging, and development tools used throughout the application.

## 📋 Table of Contents

- [Overview](#overview)
- [Configuration Management](#configuration-management)
- [Logging System](#logging-system)
- [Swagger Integration](#swagger-integration)
- [Helper Functions](#helper-functions)
- [Adding New Utilities](#adding-new-utilities)
- [Best Practices](#best-practices)

## Overview

The utilities package provides:
- **Configuration Management**: Environment-aware configuration loading with Viper
- **Logging System**: Structured logging with multiple formats and levels
- **Swagger Documentation**: Interactive API documentation with authentication
- **Security Helpers**: Password hashing and UUID generation
- **JSON Utilities**: Pretty printing and data formatting
- **Development Tools**: Debugging and development assistance

## Configuration Management

### Configuration Loading

**File**: `utils/utils.go:16-81`

The configuration system uses Viper for flexible configuration management with support for:
- JSON configuration files
- Environment variables
- Default values
- Nested configuration structures

#### Main Configuration Function
```go
func GetConfig() (*models.Config, error) {
    config, err := Load()
    if err != nil {
        return nil, fmt.Errorf("error loading config: %w", err)
    }
    return config, nil
}
```

#### Configuration Loading Process
```go
func Load() (*models.Config, error) {
    v := viper.New()
    
    // Set configuration file details
    v.SetConfigName("config")
    v.SetConfigType("json")
    v.AddConfigPath(".")
    v.AddConfigPath("./configs")
    v.AddConfigPath("../")
    v.AddConfigPath("../../")
    
    // Set default values
    setDefaults(v)
    
    // Enable environment variable support
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Try to read config file
    if err := v.ReadInConfig(); err != nil {
        // Config file not found, continue with defaults and env vars
        fmt.Printf("Config file not found (%v), using defaults and environment variables\n", err)
    } else {
        fmt.Printf("Using config file: %s\n", v.ConfigFileUsed())
    }
    
    // Handle nested JSON structure from config.json
    if v.IsSet("app") {
        // Flatten nested structure for easier mapping
        flattenNestedConfig(v)
    }
    
    var config models.Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // Parse JWT expiration if it's a string
    if v.IsSet("jwt.expires_in") {
        expiresStr := v.GetString("jwt.expires_in")
        if expiresStr != "" {
            if expires, err := time.ParseDuration(expiresStr); err != nil {
                return nil, fmt.Errorf("invalid JWT expires_in format: %w", err)
            } else {
                config.JWTExpiresIn = expires
            }
        }
    }
    
    // Validate configuration
    if err := validate(&config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return &config, nil
}
```

### Default Configuration Values

**File**: `utils/utils.go:83-129`

```go
func setDefaults(v *viper.Viper) {
    // Application defaults
    v.SetDefault("app_name", "FieldFuze Backend")
    v.SetDefault("app_version", "1.0.0")
    v.SetDefault("app_env", "development")
    v.SetDefault("app_host", "0.0.0.0")
    v.SetDefault("app_port", "8081")
    
    // JWT defaults
    v.SetDefault("jwt_secret", "your-super-secret-jwt-key-change-this-in-production")
    v.SetDefault("jwt_expires_in", 30*time.Minute)
    
    // Security & Permission defaults
    v.SetDefault("graceful_permission_degradation", true)
    v.SetDefault("permission_cache_ttl_seconds", 30)
    v.SetDefault("strict_role_validation", false)
    v.SetDefault("log_permission_changes", true)
    
    // AWS defaults
    v.SetDefault("aws_region", "us-east-1")
    v.SetDefault("aws_access_key_id", "")
    v.SetDefault("aws_secret_access_key", "")
    v.SetDefault("dynamodb_endpoint", "")
    v.SetDefault("dynamodb_table_prefix", "dev")
    
    // Telnyx defaults
    v.SetDefault("telnyx_api_key", "")
    v.SetDefault("telnyx_app_id", "")
    v.SetDefault("telnyx_public_key", "")
    
    // Logging defaults
    v.SetDefault("log_level", "info")
    v.SetDefault("log_format", "json")
    
    // CORS defaults
    v.SetDefault("cors_origins", []string{"*"})
    
    // Rate limiting defaults
    v.SetDefault("rate_limit_requests_per_minute", 100)
    
    // Base Path default
    v.SetDefault("basePath", "/api/v1")
    
    // Setup tables to create
    v.SetDefault("tables", []string{"users1"})
}
```

### Configuration Validation

```go
func validate(c *models.Config) error {
    if c.JWTSecret == "your-super-secret-jwt-key-change-this-in-production" && c.AppEnv == "production" {
        return fmt.Errorf("JWT_SECRET must be set in production environment")
    }
    
    // In production, we should have AWS credentials set
    if c.AppEnv == "production" && c.AWSAccessKeyID == "" {
        fmt.Println("No AWS credentials provided, assuming IAM role is used")
    }
    
    return nil
}
```

### Configuration Structure Flattening

For compatibility with nested JSON configuration files:

```go
func flattenNestedConfig(v *viper.Viper) {
    // App section
    if v.IsSet("app.name") {
        v.Set("app_name", v.GetString("app.name"))
    }
    if v.IsSet("app.version") {
        v.Set("app_version", v.GetString("app.version"))
    }
    // ... more flattening for other sections
    
    // JWT section
    if v.IsSet("jwt.secret") {
        v.Set("jwt_secret", v.GetString("jwt.secret"))
    }
    if v.IsSet("jwt.expires_in") {
        v.Set("jwt_expires_in", v.GetString("jwt.expires_in"))
    }
    
    // AWS section
    if v.IsSet("aws.region") {
        v.Set("aws_region", v.GetString("aws.region"))
    }
    // ... more sections
}
```

### Configuration File Examples

#### JSON Configuration (config.json)
```json
{
    "app": {
        "name": "FieldFuze Backend",
        "version": "1.0.0",
        "env": "development",
        "host": "0.0.0.0",
        "port": "8081"
    },
    "jwt": {
        "secret": "your-super-secret-jwt-key",
        "expires_in": "30m"
    },
    "aws": {
        "region": "us-east-1",
        "access_key_id": "your-access-key",
        "secret_access_key": "your-secret-key",
        "dynamodb_endpoint": "",
        "dynamodb_table_prefix": "dev"
    },
    "security": {
        "graceful_permission_degradation": true,
        "permission_cache_ttl_seconds": 30,
        "strict_role_validation": false,
        "log_permission_changes": true
    },
    "logging": {
        "level": "info",
        "format": "json"
    },
    "cors": {
        "origins": ["*"]
    },
    "rate_limit": {
        "requests_per_minute": 100
    },
    "basePath": "/api/v1/auth",
    "tables": ["users"]
}
```

#### Environment Variables
```bash
# Application
export APP_NAME="FieldFuze Backend"
export APP_VERSION="1.0.0"
export APP_ENV="production"
export APP_HOST="0.0.0.0"
export APP_PORT="8080"

# JWT
export JWT_SECRET="super-secure-production-secret"
export JWT_EXPIRES_IN="30m"

# AWS
export AWS_REGION="us-west-2"
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export DYNAMODB_TABLE_PREFIX="prod"

# Logging
export LOG_LEVEL="warn"
export LOG_FORMAT="json"
```

## Logging System

### Logger Interface

**File**: `utils/logger/logger.go:9-21`

```go
type Logger interface {
    Debug(args ...interface{})
    Debugf(format string, args ...interface{})
    Info(args ...interface{})
    Infof(format string, args ...interface{})
    Warn(args ...interface{})
    Warnf(format string, args ...interface{})
    Error(args ...interface{})
    Errorf(format string, args ...interface{})
    Fatal(args ...interface{})
    Fatalf(format string, args ...interface{})
}
```

### Logger Implementation

**File**: `utils/logger/logger.go:23-62`

```go
type LogrusLogger struct {
    logger *logrus.Logger
}

func NewLogger(level, format string) Logger {
    logger := logrus.New()
    
    // Set log level
    switch level {
    case "debug":
        logger.SetLevel(logrus.DebugLevel)
    case "info":
        logger.SetLevel(logrus.InfoLevel)
    case "warn":
        logger.SetLevel(logrus.WarnLevel)
    case "error":
        logger.SetLevel(logrus.ErrorLevel)
    default:
        logger.SetLevel(logrus.InfoLevel)
    }
    
    // Set log format
    if format == "json" {
        logger.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: "2006-01-02 15:04:05",
        })
    } else {
        logger.SetFormatter(&logrus.TextFormatter{
            FullTimestamp:   true,
            TimestampFormat: "2006-01-02 15:04:05",
            ForceColors:     true, // Force colors even when not in terminal
        })
    }
    
    logger.SetOutput(os.Stdout)
    
    return &LogrusLogger{logger: logger}
}
```

### Logger Usage Examples

#### Basic Logging
```go
logger := logger.NewLogger("info", "json")

logger.Info("Server starting...")
logger.Infof("Server listening on port %s", config.AppPort)
logger.Warn("Configuration file not found, using defaults")
logger.Error("Database connection failed", err)
logger.Fatal("Critical error, shutting down")
```

#### Structured Logging (JSON Format)
```json
{
  "level": "info",
  "msg": "User created successfully",
  "time": "2024-01-15T10:30:45Z",
  "user_id": "user-123",
  "email": "user@example.com"
}
```

#### Text Format Output
```
INFO[2024-01-15 10:30:45] User created successfully  user_id=user-123 email=user@example.com
```

### Custom Logger Wrapper

For application-specific logging with context:

```go
type ContextLogger struct {
    logger logger.Logger
    fields map[string]interface{}
}

func NewContextLogger(baseLogger logger.Logger) *ContextLogger {
    return &ContextLogger{
        logger: baseLogger,
        fields: make(map[string]interface{}),
    }
}

func (l *ContextLogger) WithField(key string, value interface{}) *ContextLogger {
    newFields := make(map[string]interface{})
    for k, v := range l.fields {
        newFields[k] = v
    }
    newFields[key] = value
    
    return &ContextLogger{
        logger: l.logger,
        fields: newFields,
    }
}

func (l *ContextLogger) WithFields(fields map[string]interface{}) *ContextLogger {
    newFields := make(map[string]interface{})
    for k, v := range l.fields {
        newFields[k] = v
    }
    for k, v := range fields {
        newFields[k] = v
    }
    
    return &ContextLogger{
        logger: l.logger,
        fields: newFields,
    }
}

func (l *ContextLogger) Info(msg string) {
    if len(l.fields) > 0 {
        l.logger.Infof("%s %+v", msg, l.fields)
    } else {
        l.logger.Info(msg)
    }
}
```

## Swagger Integration

### Swagger Configuration

**File**: `utils/swagger/swagger.go:10-14`

```go
type SwaggerConfig struct {
    Title         string
    SwaggerDocURL string
    AuthURL       string
}
```

### Enhanced Swagger UI

The Swagger integration provides:
- **Interactive API Documentation**: Full OpenAPI/Swagger UI
- **Built-in Authentication**: Login form integrated into the authorization dialog
- **Auto-token Management**: Automatic Bearer token insertion
- **Custom Styling**: Enhanced UI with better UX

#### Swagger Handler

**File**: `utils/swagger/swagger.go:262-283`

```go
func ServeSwaggerUI(config SwaggerConfig) gin.HandlerFunc {
    // Set defaults
    if config.Title == "" {
        config.Title = "API Documentation"
    }
    if config.SwaggerDocURL == "" {
        config.SwaggerDocURL = "/swagger/doc.json"
    }
    if config.AuthURL == "" {
        config.AuthURL = "/api/v1/auth/user/login"
    }
    
    tmpl := template.Must(template.New("swagger").Parse(swaggerHTML))
    
    return func(c *gin.Context) {
        c.Header("Content-Type", "text/html; charset=utf-8")
        if err := tmpl.Execute(c.Writer, config); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render Swagger UI"})
        }
    }
}
```

#### Enhanced Features

The Swagger UI includes:

1. **Integrated Login Form**: Automatically appears in authorization dialogs
2. **Token Auto-Fill**: Bearer tokens are automatically inserted after login
3. **Modern UI**: Clean, responsive design with proper styling
4. **Error Handling**: User-friendly error messages for authentication failures
5. **Real-time Monitoring**: Monitors for new authorization dialogs

#### Authentication Integration

```javascript
// Global authentication function
window.performAuthentication = async function() {
    const username = document.getElementById('login-username')?.value?.trim();
    const password = document.getElementById('login-password')?.value?.trim();
    const button = document.querySelector('.login-form-button');

    if (!username || !password) {
        alert('Please enter both username and password');
        return;
    }

    // Disable button during request
    button.disabled = true;
    button.textContent = 'Logging in...';

    try {
        const response = await fetch(window.AUTH_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: username,
                password: password
            })
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || 'Authentication failed');
        }

        const data = await response.json();
        const accessToken = data.data?.access_token;

        if (!accessToken) {
            throw new Error('No access token received');
        }

        // Find and populate the Bearer token input field
        const tokenInput = document.querySelector('input[data-name="Authorization"]');
        if (tokenInput) {
            tokenInput.value = 'Bearer ' + accessToken;
            
            // Trigger change events to notify Swagger UI
            tokenInput.dispatchEvent(new Event('input', { bubbles: true }));
            tokenInput.dispatchEvent(new Event('change', { bubbles: true }));
            
            alert('✅ Authentication successful! Bearer token has been automatically filled.');
        } else {
            // Fallback: show token to user
            alert('✅ Authentication successful!\\n\\nToken: Bearer ' + accessToken + '\\n\\nPlease paste this token in the Authorization field.');
        }

    } catch (error) {
        console.error('Authentication error:', error);
        alert('❌ Authentication failed: ' + error.message);
    } finally {
        // Re-enable button
        button.disabled = false;
        button.textContent = 'Login';
    }
};
```

#### Usage in Routes

```go
// In controller/controller.go
swaggerConfig := swagger.SwaggerConfig{
    Title:         "FieldFuze Backend API",
    SwaggerDocURL: "/swagger/doc.json",
    AuthURL:       "/api/v1/auth/user/login",
}
r.GET("/swagger", swagger.ServeSwaggerUI(swaggerConfig))
r.GET("/swagger/", swagger.ServeSwaggerUI(swaggerConfig))
r.GET("/swagger/index.html", swagger.ServeSwaggerUI(swaggerConfig))

// Swagger JSON spec
r.GET("/swagger/doc.json", func(c *gin.Context) {
    c.Header("Content-Type", "application/json")
    c.File("./docs/swagger.json")
})
```

## Helper Functions

### UUID Generation

**File**: `utils/utils.go:249-252`

```go
func GenerateUUID() string {
    return uuid.New().String()
}
```

### Password Security

**File**: `utils/utils.go:254-267`

```go
// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedBytes), nil
}

// CheckPassword compares a hashed password with a plain text password
func CheckPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

### JSON Utilities

**File**: `utils/utils.go:239-247`

```go
// PrintPrettyJSON takes any struct or map and prints it as pretty JSON
func PrintPrettyJSON(data interface{}) string {
    prettyJSON, err := json.MarshalIndent(data, "", "    ") // 4 spaces indent
    if err != nil {
        fmt.Println("Failed to generate JSON:", err)
        return ""
    }
    return string(prettyJSON)
}
```

### Usage Examples

```go
// Generate UUID for new entities
userID := utils.GenerateUUID()

// Hash password for storage
hashedPassword, err := utils.HashPassword("plainTextPassword")
if err != nil {
    log.Error("Failed to hash password:", err)
}

// Verify password during login
isValid := utils.CheckPassword(storedHash, providedPassword)

// Pretty print for debugging
fmt.Println("Config:", utils.PrintPrettyJSON(config))
fmt.Println("User Data:", utils.PrintPrettyJSON(userData))
```

## Adding New Utilities

### Step 1: Create Utility File

```go
// utils/validation.go
package utils

import (
    "net/mail"
    "regexp"
    "strings"
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

// ValidatePhoneNumber validates phone number format
func ValidatePhoneNumber(phone string) bool {
    // Remove common separators
    cleaned := strings.ReplaceAll(phone, " ", "")
    cleaned = strings.ReplaceAll(cleaned, "-", "")
    cleaned = strings.ReplaceAll(cleaned, "(", "")
    cleaned = strings.ReplaceAll(cleaned, ")", "")
    
    // Basic pattern matching
    phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
    return phoneRegex.MatchString(cleaned)
}

// SanitizeString removes dangerous characters
func SanitizeString(input string) string {
    // Remove SQL injection attempts
    sqlRegex := regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`)
    cleaned := sqlRegex.ReplaceAllString(input, "")
    
    // Remove script tags
    scriptRegex := regexp.MustCompile(`(?i)<script.*?>.*?</script>`)
    cleaned = scriptRegex.ReplaceAllString(cleaned, "")
    
    return strings.TrimSpace(cleaned)
}
```

### Step 2: Add Crypto Utilities

```go
// utils/crypto.go
package utils

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "fmt"
)

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// HashSHA256 creates a SHA256 hash of the input
func HashSHA256(input string) string {
    hash := sha256.Sum256([]byte(input))
    return hex.EncodeToString(hash[:])
}

// GenerateAPIKey generates a secure API key
func GenerateAPIKey() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return fmt.Sprintf("ff_%s", hex.EncodeToString(bytes)), nil
}
```

### Step 3: Add Time Utilities

```go
// utils/time.go
package utils

import (
    "time"
)

// TimeAgo returns a human-readable time duration
func TimeAgo(t time.Time) string {
    duration := time.Since(t)
    
    switch {
    case duration < time.Minute:
        return "just now"
    case duration < time.Hour:
        minutes := int(duration.Minutes())
        if minutes == 1 {
            return "1 minute ago"
        }
        return fmt.Sprintf("%d minutes ago", minutes)
    case duration < 24*time.Hour:
        hours := int(duration.Hours())
        if hours == 1 {
            return "1 hour ago"
        }
        return fmt.Sprintf("%d hours ago", hours)
    case duration < 30*24*time.Hour:
        days := int(duration.Hours() / 24)
        if days == 1 {
            return "1 day ago"
        }
        return fmt.Sprintf("%d days ago", days)
    default:
        return t.Format("2006-01-02")
    }
}

// StartOfDay returns the start of the day for the given time
func StartOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for the given time
func EndOfDay(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// IsWeekend checks if the given time is a weekend
func IsWeekend(t time.Time) bool {
    weekday := t.Weekday()
    return weekday == time.Saturday || weekday == time.Sunday
}
```

### Step 4: Add HTTP Utilities

```go
// utils/http.go
package utils

import (
    "fieldfuze-backend/models"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// RespondWithSuccess creates a standardized success response
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
    c.JSON(statusCode, models.APIResponse{
        Status:  "success",
        Code:    statusCode,
        Message: message,
        Data:    data,
    })
}

// RespondWithError creates a standardized error response
func RespondWithError(c *gin.Context, statusCode int, message string, errorType string, details string) {
    c.JSON(statusCode, models.APIResponse{
        Status:  "error",
        Code:    statusCode,
        Message: message,
        Error: &models.APIError{
            Type:    errorType,
            Details: details,
        },
    })
}

// RespondWithValidationError creates a validation error response
func RespondWithValidationError(c *gin.Context, message string, field string, details string) {
    c.JSON(http.StatusBadRequest, models.APIResponse{
        Status:  "error",
        Code:    http.StatusBadRequest,
        Message: message,
        Error: &models.APIError{
            Type:    "ValidationError",
            Field:   field,
            Details: details,
        },
    })
}

// GetClientIP extracts the real client IP from request
func GetClientIP(c *gin.Context) string {
    // Check X-Forwarded-For header
    if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
        return strings.Split(ip, ",")[0]
    }
    
    // Check X-Real-IP header
    if ip := c.GetHeader("X-Real-IP"); ip != "" {
        return ip
    }
    
    // Fall back to RemoteAddr
    return c.ClientIP()
}
```

## Best Practices

### 1. Configuration Management
- **Environment Variables**: Use environment variables for sensitive data
- **Default Values**: Provide sensible defaults for all configuration
- **Validation**: Validate configuration at startup
- **Documentation**: Document all configuration options

### 2. Logging Standards
- **Structured Logging**: Use structured logging for better searchability
- **Log Levels**: Use appropriate log levels (Debug, Info, Warn, Error, Fatal)
- **Context**: Include relevant context in log messages
- **Performance**: Don't log in hot paths unless necessary

### 3. Error Handling
- **Consistent Responses**: Use standardized error response formats
- **Error Context**: Provide helpful error messages and context
- **Security**: Don't leak sensitive information in error messages
- **Logging**: Log errors with sufficient detail for debugging

### 4. Security
- **Input Validation**: Validate and sanitize all inputs
- **Password Security**: Use bcrypt for password hashing
- **Secrets Management**: Never hardcode secrets in source code
- **HTTPS**: Use HTTPS in production

### 5. Performance
- **Efficient JSON**: Use efficient JSON marshaling/unmarshaling
- **Caching**: Cache expensive operations where appropriate
- **Memory Usage**: Be mindful of memory allocations
- **Concurrency**: Use appropriate concurrency patterns

### 6. Testing
- **Unit Tests**: Write tests for utility functions
- **Test Coverage**: Maintain good test coverage
- **Edge Cases**: Test edge cases and error conditions
- **Mocking**: Use mocks for external dependencies

### 7. Documentation
- **Code Comments**: Document complex logic and algorithms
- **API Documentation**: Keep API documentation up to date
- **Examples**: Provide usage examples for utilities
- **README**: Maintain comprehensive README files

---

**Related Documentation**: [Configuration](CONFIGURATION.md) | [Development Guide](DEVELOPMENT.md) | [API Reference](API.md) | [Security](SECURITY.md)