package utils

import (
	"encoding/json"
	"fieldfuze-backend/models"

	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// GetConfig read the configuration from environment variables or config files
func GetConfig() (*models.Config, error) {
	config, err := Load()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}
	return config, nil
}

// Load initializes and returns the application configuration using Viper
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

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Application defaults
	v.SetDefault("app_name", "FieldFuze Backend")
	v.SetDefault("app_version", "1.0.0")
	v.SetDefault("app_env", "development")
	v.SetDefault("app_host", "0.0.0.0")
	v.SetDefault("app_port", "8081")

	// JWT defaults
	v.SetDefault("jwt_secret", "your-super-secret-jwt-key-change-this-in-production")
	v.SetDefault("jwt_expires_in", 30*time.Minute) // Shorter token expiration for better security

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

	// setup tables to create
	v.SetDefault("tables", []string{"users1"})
}

// validate checks if all required configuration is provided
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

// flattenNestedConfig flattens the nested JSON structure to flat keys for easier mapping
func flattenNestedConfig(v *viper.Viper) {
	// App section
	if v.IsSet("app.name") {
		v.Set("app_name", v.GetString("app.name"))
	}
	if v.IsSet("app.version") {
		v.Set("app_version", v.GetString("app.version"))
	}
	if v.IsSet("app.env") {
		v.Set("app_env", v.GetString("app.env"))
	}
	if v.IsSet("app.host") {
		v.Set("app_host", v.GetString("app.host"))
	}
	if v.IsSet("app.port") {
		v.Set("app_port", v.GetString("app.port"))
	}

	// JWT section
	if v.IsSet("jwt.secret") {
		v.Set("jwt_secret", v.GetString("jwt.secret"))
	}

	// AWS section
	if v.IsSet("aws.region") {
		v.Set("aws_region", v.GetString("aws.region"))
	}
	if v.IsSet("aws.access_key_id") {
		v.Set("aws_access_key_id", v.GetString("aws.access_key_id"))
	}
	if v.IsSet("aws.secret_access_key") {
		v.Set("aws_secret_access_key", v.GetString("aws.secret_access_key"))
	}
	if v.IsSet("aws.dynamodb_endpoint") {
		v.Set("dynamodb_endpoint", v.GetString("aws.dynamodb_endpoint"))
	}
	if v.IsSet("aws.dynamodb_table_prefix") {
		v.Set("dynamodb_table_prefix", v.GetString("aws.dynamodb_table_prefix"))
	}

	// Telnyx section
	if v.IsSet("telnyx.api_key") {
		v.Set("telnyx_api_key", v.GetString("telnyx.api_key"))
	}
	if v.IsSet("telnyx.app_id") {
		v.Set("telnyx_app_id", v.GetString("telnyx.app_id"))
	}
	if v.IsSet("telnyx.public_key") {
		v.Set("telnyx_public_key", v.GetString("telnyx.public_key"))
	}

	// Logging section
	if v.IsSet("logging.level") {
		v.Set("log_level", v.GetString("logging.level"))
	}
	if v.IsSet("logging.format") {
		v.Set("log_format", v.GetString("logging.format"))
	}

	// CORS section
	if v.IsSet("cors.origins") {
		v.Set("cors_origins", v.GetStringSlice("cors.origins"))
	}

	// Rate limit section
	if v.IsSet("rate_limit.requests_per_minute") {
		v.Set("rate_limit_requests_per_minute", v.GetInt("rate_limit.requests_per_minute"))
	}

	// Base Path
	if v.IsSet("basePath") {
		v.Set("basePath", v.GetString("basePath"))
	}
}

// PrintPrettyJSON takes any struct or map and prints it as pretty JSON
func PrintPrettyJSON(data interface{}) string {
	prettyJSON, err := json.MarshalIndent(data, "", "    ") // 4 spaces indent
	if err != nil {
		fmt.Println("Failed to generate JSON:", err)
		return ""
	}
	return string(prettyJSON)
}

// GenerateUUID returns a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPassword compares a hashed password with a plain text password.
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
