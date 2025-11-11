package worker

import (
	"fieldfuze-backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLockManager(t *testing.T) {
	lockPath := "/tmp/test.lock"
	timeout := 5 * time.Minute
	env := "test"

	manager := NewLockManager(lockPath, timeout, env)

	assert.NotNil(t, manager)
	assert.Equal(t, lockPath, manager.LockFilePath)
	assert.Equal(t, timeout, manager.LockTimeout)
	assert.Equal(t, env, manager.Environment)
}

func TestLockManagerStructure(t *testing.T) {
	// Test basic structure
	manager := &LockManager{}

	assert.NotNil(t, manager)

	// Test that we can access the embedded struct
	assert.NotPanics(t, func() {
		_ = manager.LockManager
	})
}

func TestLockManagerConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		lockPath string
		timeout  time.Duration
		env      string
	}{
		{
			name:     "Valid configuration",
			lockPath: "/tmp/valid.lock",
			timeout:  5 * time.Minute,
			env:      "test",
		},
		{
			name:     "Short timeout",
			lockPath: "/tmp/short.lock",
			timeout:  30 * time.Second,
			env:      "dev",
		},
		{
			name:     "Long timeout",
			lockPath: "/tmp/long.lock",
			timeout:  30 * time.Minute,
			env:      "prod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewLockManager(tt.lockPath, tt.timeout, tt.env)

			assert.NotNil(t, manager)
			assert.Equal(t, tt.lockPath, manager.LockFilePath)
			assert.Equal(t, tt.timeout, manager.LockTimeout)
			assert.Equal(t, tt.env, manager.Environment)
		})
	}
}

func TestLockManagerFields(t *testing.T) {
	lockPath := "/tmp/test.lock"
	timeout := 5 * time.Minute
	env := "test"

	manager := NewLockManager(lockPath, timeout, env)

	// Test field access
	assert.NotPanics(t, func() {
		_ = manager.LockFilePath
		_ = manager.LockTimeout
		_ = manager.Environment
	})

	// Test that fields are correctly set
	assert.Equal(t, lockPath, manager.LockFilePath)
	assert.Equal(t, timeout, manager.LockTimeout)
	assert.Equal(t, env, manager.Environment)
}

func TestLockManagerTypes(t *testing.T) {
	manager := NewLockManager("/tmp/test.lock", 5*time.Minute, "test")

	// Test that types are correct
	assert.IsType(t, &models.LockManager{}, manager)
	assert.IsType(t, "", manager.LockFilePath)
	assert.IsType(t, time.Duration(0), manager.LockTimeout)
	assert.IsType(t, "", manager.Environment)
}

func TestLockManagerWithDifferentPaths(t *testing.T) {
	paths := []string{
		"/tmp/test1.lock",
		"/var/tmp/test2.lock",
		"./test3.lock",
		"test4.lock",
	}

	for _, path := range paths {
		manager := NewLockManager(path, time.Minute, "test")

		assert.NotNil(t, manager)
		assert.Equal(t, path, manager.LockFilePath)
	}
}
