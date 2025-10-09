package utils

import (
	"encoding/json"
	"fieldfuze-backend/models"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

// UtilsTestSuite defines a test suite for utils functions
type UtilsTestSuite struct {
	suite.Suite
	originalEnv map[string]string
}

// SetupTest runs before each test
func (suite *UtilsTestSuite) SetupTest() {
	// Store original environment variables
	suite.originalEnv = make(map[string]string)
	envVars := []string{
		"APP_NAME", "APP_VERSION", "APP_ENV", "APP_HOST", "APP_PORT",
		"JWT_SECRET", "JWT_EXPIRES_IN",
		"AWS_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"DYNAMODB_ENDPOINT", "DYNAMODB_TABLE_PREFIX",
		"TELNYX_API_KEY", "TELNYX_APP_ID", "TELNYX_PUBLIC_KEY",
		"LOG_LEVEL", "LOG_FORMAT",
		"CORS_ORIGINS", "RATE_LIMIT_REQUESTS_PER_MINUTE",
		"BASEPATH",
	}
	
	for _, envVar := range envVars {
		suite.originalEnv[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}
}

// TearDownTest runs after each test
func (suite *UtilsTestSuite) TearDownTest() {
	// Restore original environment variables
	for envVar, value := range suite.originalEnv {
		if value != "" {
			os.Setenv(envVar, value)
		} else {
			os.Unsetenv(envVar)
		}
	}
}

// TestGetConfig tests the GetConfig function
func (suite *UtilsTestSuite) TestGetConfig() {
	config, err := GetConfig()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	
	// Test default values
	assert.Equal(suite.T(), "FieldFuze Backend", config.AppName)
	assert.Equal(suite.T(), "1.0.0", config.AppVersion)
	assert.Equal(suite.T(), "development", config.AppEnv)
	assert.Equal(suite.T(), "0.0.0.0", config.AppHost)
	assert.Equal(suite.T(), "8081", config.AppPort)
}

// TestGetConfigWithEnvironmentVariables tests GetConfig with environment variables
func (suite *UtilsTestSuite) TestGetConfigWithEnvironmentVariables() {
	// Set environment variables
	os.Setenv("APP_NAME", "Test App")
	os.Setenv("APP_VERSION", "2.0.0")
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "production-secret")
	os.Setenv("AWS_REGION", "us-west-2")
	
	config, err := GetConfig()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	
	// Test environment variable overrides
	assert.Equal(suite.T(), "Test App", config.AppName)
	assert.Equal(suite.T(), "2.0.0", config.AppVersion)
	assert.Equal(suite.T(), "production", config.AppEnv)
	assert.Equal(suite.T(), "production-secret", config.JWTSecret)
	assert.Equal(suite.T(), "us-west-2", config.AWSRegion)
}

// TestLoad tests the Load function with defaults
func (suite *UtilsTestSuite) TestLoad() {
	config, err := Load()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	
	// Verify config file values (config.json overrides defaults)
	assert.Equal(suite.T(), "FieldFuze Backend", config.AppName)
	assert.Equal(suite.T(), "1.0.0", config.AppVersion)
	assert.Equal(suite.T(), "development", config.AppEnv)
	assert.Equal(suite.T(), "0.0.0.0", config.AppHost)
	assert.Equal(suite.T(), "8081", config.AppPort)
	assert.Equal(suite.T(), "your-super-secret-jwt-key-change-this-in-production", config.JWTSecret)
	assert.Equal(suite.T(), 24*time.Hour, config.JWTExpiresIn) // From config.json: "24h"
	assert.Equal(suite.T(), "us-east-1", config.AWSRegion)
	assert.Equal(suite.T(), "debug", config.LogLevel) // From config.json: "debug"
	assert.Equal(suite.T(), "text", config.LogFormat) // From config.json: "text"
	assert.Equal(suite.T(), []string{"*"}, config.CORSOrigins)
	assert.Equal(suite.T(), 100, config.RateLimitRequestsPerMinute)
	assert.Equal(suite.T(), "/api/v1/auth", config.BasePath) // From config.json: "/api/v1/auth"
	assert.Equal(suite.T(), []string{"users1", "role", "organization"}, config.Tables) // From config.json
}

// TestLoadWithJWTExpirationString tests JWT expiration parsing
func (suite *UtilsTestSuite) TestLoadWithJWTExpirationString() {
	os.Setenv("JWT_EXPIRES_IN", "24h")
	
	config, err := Load()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 24*time.Hour, config.JWTExpiresIn)
}

// TestLoadWithInvalidJWTExpiration tests invalid JWT expiration
func (suite *UtilsTestSuite) TestLoadWithInvalidJWTExpiration() {
	os.Setenv("JWT_EXPIRES_IN", "invalid-duration")
	
	config, err := Load()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	// The error message might be different depending on which parsing fails
	assert.True(suite.T(), strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "failed"))
}

// TestLoadWithProductionValidation tests production environment validation
func (suite *UtilsTestSuite) TestLoadWithProductionValidation() {
	os.Setenv("APP_ENV", "production")
	// Don't set JWT_SECRET, should use default which should fail validation
	
	config, err := Load()
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "JWT_SECRET must be set in production environment")
}

// TestLoadWithProductionValidationWithValidSecret tests production with valid secret
func (suite *UtilsTestSuite) TestLoadWithProductionValidationWithValidSecret() {
	os.Setenv("APP_ENV", "production")
	os.Setenv("JWT_SECRET", "production-secret-key")
	
	config, err := Load()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), "production", config.AppEnv)
	assert.Equal(suite.T(), "production-secret-key", config.JWTSecret)
}

// TestValidate tests the validate function
func (suite *UtilsTestSuite) TestValidate() {
	// Test development environment - should pass
	config := &models.Config{
		AppEnv:    "development",
		JWTSecret: "your-super-secret-jwt-key-change-this-in-production",
	}
	err := validate(config)
	assert.NoError(suite.T(), err)
}

// TestValidateProductionWithDefaultSecret tests validation failure in production
func (suite *UtilsTestSuite) TestValidateProductionWithDefaultSecret() {
	config := &models.Config{
		AppEnv:    "production",
		JWTSecret: "your-super-secret-jwt-key-change-this-in-production",
	}
	err := validate(config)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "JWT_SECRET must be set in production environment")
}

// TestValidateProductionWithValidSecret tests validation success in production
func (suite *UtilsTestSuite) TestValidateProductionWithValidSecret() {
	config := &models.Config{
		AppEnv:    "production",
		JWTSecret: "production-secret",
	}
	err := validate(config)
	assert.NoError(suite.T(), err)
}

// TestValidateProductionWithNoAWSCredentials tests production without AWS credentials
func (suite *UtilsTestSuite) TestValidateProductionWithNoAWSCredentials() {
	config := &models.Config{
		AppEnv:          "production",
		JWTSecret:       "production-secret",
		AWSAccessKeyID:  "",
	}
	err := validate(config)
	assert.NoError(suite.T(), err) // Should not error, just print warning
}

// TestPrintPrettyJSON tests the PrintPrettyJSON function
func (suite *UtilsTestSuite) TestPrintPrettyJSON() {
	data := map[string]interface{}{
		"name":  "test",
		"value": 123,
		"array": []string{"a", "b", "c"},
	}
	
	result := PrintPrettyJSON(data)
	assert.NotEmpty(suite.T(), result)
	
	// Verify it's valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(result), &parsed)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", parsed["name"])
	assert.Equal(suite.T(), float64(123), parsed["value"]) // JSON numbers are float64
}

// TestPrintPrettyJSONWithNil tests PrintPrettyJSON with nil input
func (suite *UtilsTestSuite) TestPrintPrettyJSONWithNil() {
	result := PrintPrettyJSON(nil)
	assert.Equal(suite.T(), "null", result)
}

// TestPrintPrettyJSONWithStruct tests PrintPrettyJSON with struct
func (suite *UtilsTestSuite) TestPrintPrettyJSONWithStruct() {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	
	data := TestStruct{Name: "test", Value: 42}
	result := PrintPrettyJSON(data)
	assert.NotEmpty(suite.T(), result)
	assert.Contains(suite.T(), result, "\"name\": \"test\"")
	assert.Contains(suite.T(), result, "\"value\": 42")
}

// TestPrintPrettyJSONWithInvalidData tests PrintPrettyJSON with non-serializable data
func (suite *UtilsTestSuite) TestPrintPrettyJSONWithInvalidData() {
	// Create a channel which cannot be marshaled to JSON
	invalidData := make(chan int)
	result := PrintPrettyJSON(invalidData)
	assert.Empty(suite.T(), result)
}

// TestGenerateUUID tests the GenerateUUID function
func (suite *UtilsTestSuite) TestGenerateUUID() {
	// Test basic generation
	id1 := GenerateUUID()
	id2 := GenerateUUID()
	
	assert.NotEmpty(suite.T(), id1)
	assert.NotEmpty(suite.T(), id2)
	assert.NotEqual(suite.T(), id1, id2)
	
	// Test UUID format
	_, err := uuid.Parse(id1)
	assert.NoError(suite.T(), err)
	
	_, err = uuid.Parse(id2)
	assert.NoError(suite.T(), err)
}

// TestGenerateUUIDUniqueness tests UUID uniqueness
func (suite *UtilsTestSuite) TestGenerateUUIDUniqueness() {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateUUID()
		assert.False(suite.T(), seen[id], "Generated duplicate UUID: %s", id)
		seen[id] = true
	}
}

// TestHashPassword tests the HashPassword function
func (suite *UtilsTestSuite) TestHashPassword() {
	password := "testpassword123"
	
	hash1, err := HashPassword(password)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hash1)
	assert.NotEqual(suite.T(), password, hash1)
	
	// Generate another hash for the same password
	hash2, err := HashPassword(password)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hash2)
	assert.NotEqual(suite.T(), hash1, hash2) // bcrypt generates different hashes each time
}

// TestHashPasswordWithEmptyString tests HashPassword with empty string
func (suite *UtilsTestSuite) TestHashPasswordWithEmptyString() {
	hash, err := HashPassword("")
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hash)
}

// TestHashPasswordWithSpecialCharacters tests HashPassword with special characters
func (suite *UtilsTestSuite) TestHashPasswordWithSpecialCharacters() {
	passwords := []string{
		"!@#$%^&*()",
		"password with spaces",
		"påssw0rd",
		"密码",
		"مرور",
	}
	
	for _, password := range passwords {
		hash, err := HashPassword(password)
		assert.NoError(suite.T(), err, "Failed to hash password: %s", password)
		assert.NotEmpty(suite.T(), hash)
		assert.NotEqual(suite.T(), password, hash)
	}
}

// TestCheckPassword tests the CheckPassword function
func (suite *UtilsTestSuite) TestCheckPassword() {
	password := "testpassword123"
	
	hash, err := HashPassword(password)
	require.NoError(suite.T(), err)
	
	// Test correct password
	assert.True(suite.T(), CheckPassword(hash, password))
	
	// Test incorrect password
	assert.False(suite.T(), CheckPassword(hash, "wrongpassword"))
	assert.False(suite.T(), CheckPassword(hash, ""))
	assert.False(suite.T(), CheckPassword(hash, "TESTPASSWORD123"))
}

// TestCheckPasswordWithInvalidHash tests CheckPassword with invalid hash
func (suite *UtilsTestSuite) TestCheckPasswordWithInvalidHash() {
	assert.False(suite.T(), CheckPassword("invalid-hash", "password"))
	assert.False(suite.T(), CheckPassword("", "password"))
	assert.False(suite.T(), CheckPassword("not-a-bcrypt-hash", "password"))
}

// TestCheckPasswordWithEmptyInputs tests CheckPassword with empty inputs
func (suite *UtilsTestSuite) TestCheckPasswordWithEmptyInputs() {
	hash, err := HashPassword("password")
	require.NoError(suite.T(), err)
	
	assert.False(suite.T(), CheckPassword(hash, ""))
	assert.False(suite.T(), CheckPassword("", "password"))
	assert.False(suite.T(), CheckPassword("", ""))
}

// TestCheckPasswordWithBcryptGenerated tests CheckPassword with bcrypt-generated hash
func (suite *UtilsTestSuite) TestCheckPasswordWithBcryptGenerated() {
	password := "testpassword"
	
	// Generate hash using bcrypt directly
	directHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(suite.T(), err)
	
	// Test with our CheckPassword function
	assert.True(suite.T(), CheckPassword(string(directHash), password))
	assert.False(suite.T(), CheckPassword(string(directHash), "wrongpassword"))
}

// Run the test suite
func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

// Standalone tests for edge cases

func TestHashPasswordEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"Very short password", "a"},
		{"Long password", strings.Repeat("a", 70)}, // bcrypt has 72 byte limit
		{"Unicode password", "test🔐password"},
		{"Password with null bytes", "test\x00password"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.True(t, CheckPassword(hash, tc.password))
		})
	}
}

func TestPasswordConsistency(t *testing.T) {
	password := "consistencytest"
	
	// Generate hash
	hash, err := HashPassword(password)
	require.NoError(t, err)
	
	// Test multiple times to ensure consistency
	for i := 0; i < 10; i++ {
		assert.True(t, CheckPassword(hash, password), "Password check failed on iteration %d", i)
	}
}

func TestUUIDFormatValidation(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := GenerateUUID()
		
		// Check UUID format (should be 36 characters with dashes)
		assert.Len(t, id, 36)
		assert.Contains(t, id, "-")
		
		// Should contain only hex characters and dashes
		for _, char := range id {
			assert.True(t, 
				(char >= '0' && char <= '9') ||
				(char >= 'a' && char <= 'f') ||
				(char >= 'A' && char <= 'F') ||
				char == '-',
				"Invalid character in UUID: %c", char)
		}
	}
}

func TestConfigValidationEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		config      *models.Config
		shouldError bool
	}{
		{
			name: "Development with default secret",
			config: &models.Config{
				AppEnv:    "development",
				JWTSecret: "your-super-secret-jwt-key-change-this-in-production",
			},
			shouldError: false,
		},
		{
			name: "Test environment with default secret",
			config: &models.Config{
				AppEnv:    "test",
				JWTSecret: "your-super-secret-jwt-key-change-this-in-production",
			},
			shouldError: false,
		},
		{
			name: "Production with custom secret",
			config: &models.Config{
				AppEnv:    "production",
				JWTSecret: "custom-production-secret",
			},
			shouldError: false,
		},
		{
			name: "Staging with default secret",
			config: &models.Config{
				AppEnv:    "staging",
				JWTSecret: "your-super-secret-jwt-key-change-this-in-production",
			},
			shouldError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.config)
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}