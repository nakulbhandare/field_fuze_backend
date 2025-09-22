package models

import "time"

// Config holds all configuration for the application
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
