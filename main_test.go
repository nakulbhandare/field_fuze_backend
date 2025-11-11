package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Save original environment
	originalConfigFile := os.Getenv("CONFIG_FILE")
	defer func() {
		if originalConfigFile == "" {
			os.Unsetenv("CONFIG_FILE")
		} else {
			os.Setenv("CONFIG_FILE", originalConfigFile)
		}
	}()

	// Set test config file path
	os.Setenv("CONFIG_FILE", "config.json")

	// This test assumes config.json exists in the current directory
	// If it doesn't exist, the test may fail, but we're testing the Init function itself
	assert.NotPanics(t, func() {
		Init()
	})

	// After Init, config should be set
	assert.NotNil(t, config)
}
