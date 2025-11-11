package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigStruct(t *testing.T) {
	config := Config{
		AppName:                       "TestApp",
		AppVersion:                    "1.0.0",
		AppEnv:                        "test",
		AppHost:                       "localhost",
		AppPort:                       "8080",
		JWTSecret:                     "secret",
		JWTExpiresIn:                  24 * time.Hour,
		GracefulPermissionDegradation: true,
		PermissionCacheTTLSeconds:     300,
		StrictRoleValidation:          false,
		LogPermissionChanges:          true,
		AWSRegion:                     "us-east-1",
		DynamoDBTablePrefix:           "test_",
		LogLevel:                      "info",
		LogFormat:                     "json",
	}

	assert.Equal(t, "TestApp", config.AppName)
	assert.Equal(t, "1.0.0", config.AppVersion)
	assert.Equal(t, "test", config.AppEnv)
	assert.Equal(t, "localhost", config.AppHost)
	assert.Equal(t, "8080", config.AppPort)
	assert.Equal(t, "secret", config.JWTSecret)
	assert.Equal(t, 24*time.Hour, config.JWTExpiresIn)
	assert.True(t, config.GracefulPermissionDegradation)
	assert.Equal(t, 300, config.PermissionCacheTTLSeconds)
	assert.False(t, config.StrictRoleValidation)
	assert.True(t, config.LogPermissionChanges)
	assert.Equal(t, "us-east-1", config.AWSRegion)
	assert.Equal(t, "test_", config.DynamoDBTablePrefix)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "json", config.LogFormat)
}

func TestUserStruct(t *testing.T) {
	now := time.Now()
	user := User{
		ID:            "user123",
		Email:         "test@example.com",
		Username:      "testuser",
		FirstName:     "Test",
		LastName:      "User",
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: true,
	}

	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test", user.FirstName)
	assert.Equal(t, "User", user.LastName)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	assert.True(t, user.EmailVerified)
}

func TestJobStruct(t *testing.T) {
	now := time.Now()
	job := Job{
		JobID:     "job123",
		ClientID:  "client123",
		JobsName:  "Test Job",
		CreatedAt: now,
		CreatedData: CreatedData{
			UID:        "user123",
			UserName:   "testuser",
			UserStatus: "active",
		},
	}

	assert.Equal(t, "job123", job.JobID)
	assert.Equal(t, "client123", job.ClientID)
	assert.Equal(t, "Test Job", job.JobsName)
	assert.Equal(t, now, job.CreatedAt)
	assert.Equal(t, "user123", job.CreatedData.UID)
}

func TestRoleStruct(t *testing.T) {
	role := Role{
		ID:          "role123",
		Name:        "admin",
		Description: "Administrator role with full access",
		Level:       10,
		Permissions: []string{"read", "write", "admin"},
	}

	assert.Equal(t, "role123", role.ID)
	assert.Equal(t, "admin", role.Name)
	assert.Equal(t, "Administrator role with full access", role.Description)
	assert.Equal(t, 10, role.Level)
	assert.Contains(t, role.Permissions, "admin")
}

func TestOrganizationStruct(t *testing.T) {
	now := time.Now()
	org := Organization{
		ID:          "org123",
		Name:        "Test Org",
		Description: "Test organization",
		Status:      "active",
		CreatedAt:   now,
	}

	assert.Equal(t, "org123", org.ID)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, "Test organization", org.Description)
	assert.Equal(t, OrganizationStatus("active"), org.Status)
	assert.Equal(t, now, org.CreatedAt)
}
