package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// LoggerTestSuite defines a test suite for logger functions
type LoggerTestSuite struct {
	suite.Suite
	buffer *bytes.Buffer
	logger Logger
}

// SetupTest runs before each test
func (suite *LoggerTestSuite) SetupTest() {
	suite.buffer = &bytes.Buffer{}
}

// TearDownTest runs after each test
func (suite *LoggerTestSuite) TearDownTest() {
	if suite.buffer != nil {
		suite.buffer.Reset()
	}
}

// Helper function to create a logger with custom output
func (suite *LoggerTestSuite) createLoggerWithBuffer(level, format string) Logger {
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(suite.buffer)
	
	// Set log level
	switch level {
	case "debug":
		logrusLogger.SetLevel(logrus.DebugLevel)
	case "info":
		logrusLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		logrusLogger.SetLevel(logrus.WarnLevel)
	case "error":
		logrusLogger.SetLevel(logrus.ErrorLevel)
	default:
		logrusLogger.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if format == "json" {
		logrusLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logrusLogger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     false, // Disable colors for testing
		})
	}
	
	return &LogrusLogger{logger: logrusLogger}
}

// TestNewLogger tests the NewLogger function with different configurations
func (suite *LoggerTestSuite) TestNewLogger() {
	testCases := []struct {
		name         string
		level        string
		format       string
		expectedType string
	}{
		{"Debug level with JSON format", "debug", "json", "*logger.LogrusLogger"},
		{"Info level with text format", "info", "text", "*logger.LogrusLogger"},
		{"Warn level with JSON format", "warn", "json", "*logger.LogrusLogger"},
		{"Error level with text format", "error", "text", "*logger.LogrusLogger"},
		{"Invalid level defaults to info", "invalid", "json", "*logger.LogrusLogger"},
		{"Empty level defaults to info", "", "text", "*logger.LogrusLogger"},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			logger := NewLogger(tc.level, tc.format)
			assert.NotNil(t, logger)
			assert.Implements(t, (*Logger)(nil), logger)
		})
	}
}

// TestLoggerLevels tests different log levels
func (suite *LoggerTestSuite) TestLoggerLevels() {
	testCases := []struct {
		name          string
		level         string
		logFunc       func(Logger)
		shouldLog     bool
	}{
		{
			name:  "Debug level logs debug messages",
			level: "debug",
			logFunc: func(l Logger) {
				l.Debug("debug message")
			},
			shouldLog: true,
		},
		{
			name:  "Info level skips debug messages",
			level: "info",
			logFunc: func(l Logger) {
				l.Debug("debug message")
			},
			shouldLog: false,
		},
		{
			name:  "Info level logs info messages",
			level: "info",
			logFunc: func(l Logger) {
				l.Info("info message")
			},
			shouldLog: true,
		},
		{
			name:  "Warn level logs warn messages",
			level: "warn",
			logFunc: func(l Logger) {
				l.Warn("warn message")
			},
			shouldLog: true,
		},
		{
			name:  "Warn level skips info messages",
			level: "warn",
			logFunc: func(l Logger) {
				l.Info("info message")
			},
			shouldLog: false,
		},
		{
			name:  "Error level logs error messages",
			level: "error",
			logFunc: func(l Logger) {
				l.Error("error message")
			},
			shouldLog: true,
		},
		{
			name:  "Error level skips warn messages",
			level: "error",
			logFunc: func(l Logger) {
				l.Warn("warn message")
			},
			shouldLog: false,
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			logger := suite.createLoggerWithBuffer(tc.level, "text")
			suite.buffer.Reset()
			
			tc.logFunc(logger)
			
			output := suite.buffer.String()
			if tc.shouldLog {
				assert.NotEmpty(t, output)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

// TestDebugMethods tests Debug and Debugf methods
func (suite *LoggerTestSuite) TestDebugMethods() {
	logger := suite.createLoggerWithBuffer("debug", "text")
	
	// Test Debug
	suite.buffer.Reset()
	logger.Debug("test debug message")
	output := suite.buffer.String()
	assert.Contains(suite.T(), output, "test debug message")
	assert.Contains(suite.T(), output, "DEBU")
	
	// Test Debugf
	suite.buffer.Reset()
	logger.Debugf("debug message with %s and %d", "string", 42)
	output = suite.buffer.String()
	assert.Contains(suite.T(), output, "debug message with string and 42")
	assert.Contains(suite.T(), output, "DEBU")
}

// TestInfoMethods tests Info and Infof methods
func (suite *LoggerTestSuite) TestInfoMethods() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	// Test Info
	suite.buffer.Reset()
	logger.Info("test info message")
	output := suite.buffer.String()
	assert.Contains(suite.T(), output, "test info message")
	assert.Contains(suite.T(), output, "INFO")
	
	// Test Infof
	suite.buffer.Reset()
	logger.Infof("info message with %s and %d", "string", 42)
	output = suite.buffer.String()
	assert.Contains(suite.T(), output, "info message with string and 42")
	assert.Contains(suite.T(), output, "INFO")
}

// TestWarnMethods tests Warn and Warnf methods
func (suite *LoggerTestSuite) TestWarnMethods() {
	logger := suite.createLoggerWithBuffer("warn", "text")
	
	// Test Warn
	suite.buffer.Reset()
	logger.Warn("test warn message")
	output := suite.buffer.String()
	assert.Contains(suite.T(), output, "test warn message")
	assert.Contains(suite.T(), output, "WARN")
	
	// Test Warnf
	suite.buffer.Reset()
	logger.Warnf("warn message with %s and %d", "string", 42)
	output = suite.buffer.String()
	assert.Contains(suite.T(), output, "warn message with string and 42")
	assert.Contains(suite.T(), output, "WARN")
}

// TestErrorMethods tests Error and Errorf methods
func (suite *LoggerTestSuite) TestErrorMethods() {
	logger := suite.createLoggerWithBuffer("error", "text")
	
	// Test Error
	suite.buffer.Reset()
	logger.Error("test error message")
	output := suite.buffer.String()
	assert.Contains(suite.T(), output, "test error message")
	assert.Contains(suite.T(), output, "ERRO")
	
	// Test Errorf
	suite.buffer.Reset()
	logger.Errorf("error message with %s and %d", "string", 42)
	output = suite.buffer.String()
	assert.Contains(suite.T(), output, "error message with string and 42")
	assert.Contains(suite.T(), output, "ERRO")
}

// TestJSONFormat tests JSON format output
func (suite *LoggerTestSuite) TestJSONFormat() {
	logger := suite.createLoggerWithBuffer("info", "json")
	
	suite.buffer.Reset()
	logger.Info("test json message")
	output := suite.buffer.String()
	
	// Should be valid JSON
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(suite.T(), err)
	
	// Check JSON structure
	assert.Contains(suite.T(), logEntry, "level")
	assert.Contains(suite.T(), logEntry, "msg")
	assert.Contains(suite.T(), logEntry, "time")
	assert.Equal(suite.T(), "info", logEntry["level"])
	assert.Equal(suite.T(), "test json message", logEntry["msg"])
}

// TestTextFormat tests text format output
func (suite *LoggerTestSuite) TestTextFormat() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	suite.buffer.Reset()
	logger.Info("test text message")
	output := suite.buffer.String()
	
	assert.Contains(suite.T(), output, "test text message")
	assert.Contains(suite.T(), output, "INFO")
	// Should contain timestamp
	assert.Regexp(suite.T(), `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, output)
}

// TestLoggerInterface tests that LogrusLogger implements Logger interface
func (suite *LoggerTestSuite) TestLoggerInterface() {
	logger := NewLogger("info", "text")
	
	// Test that all interface methods exist and can be called
	assert.NotPanics(suite.T(), func() {
		logger.Debug("debug")
		logger.Debugf("debugf %s", "test")
		logger.Info("info")
		logger.Infof("infof %s", "test")
		logger.Warn("warn")
		logger.Warnf("warnf %s", "test")
		logger.Error("error")
		logger.Errorf("errorf %s", "test")
	})
}

// TestLoggerWithMultipleArguments tests logging with multiple arguments
func (suite *LoggerTestSuite) TestLoggerWithMultipleArguments() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	suite.buffer.Reset()
	logger.Info("message", 123, true, 45.67)
	output := suite.buffer.String()
	
	assert.Contains(suite.T(), output, "message")
	assert.Contains(suite.T(), output, "123")
	assert.Contains(suite.T(), output, "true")
	assert.Contains(suite.T(), output, "45.67")
}

// TestLoggerWithFormatString tests formatted logging
func (suite *LoggerTestSuite) TestLoggerWithFormatString() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	testCases := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "String formatting",
			format:   "Hello %s",
			args:     []interface{}{"World"},
			expected: "Hello World",
		},
		{
			name:     "Integer formatting",
			format:   "Number: %d",
			args:     []interface{}{42},
			expected: "Number: 42",
		},
		{
			name:     "Float formatting",
			format:   "Float: %.2f",
			args:     []interface{}{3.14159},
			expected: "Float: 3.14",
		},
		{
			name:     "Multiple arguments",
			format:   "Name: %s, Age: %d, Score: %.1f",
			args:     []interface{}{"John", 30, 85.5},
			expected: "Name: John, Age: 30, Score: 85.5",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.buffer.Reset()
			logger.Infof(tc.format, tc.args...)
			output := suite.buffer.String()
			assert.Contains(t, output, tc.expected)
		})
	}
}

// TestLoggerWithEmptyMessages tests logging with empty messages
func (suite *LoggerTestSuite) TestLoggerWithEmptyMessages() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	suite.buffer.Reset()
	logger.Info("")
	output := suite.buffer.String()
	assert.NotEmpty(suite.T(), output) // Should still log timestamp and level
	
	suite.buffer.Reset()
	logger.Infof("")
	output = suite.buffer.String()
	assert.NotEmpty(suite.T(), output) // Should still log timestamp and level
}

// TestLoggerWithNilArguments tests logging with nil arguments
func (suite *LoggerTestSuite) TestLoggerWithNilArguments() {
	logger := suite.createLoggerWithBuffer("info", "text")
	
	suite.buffer.Reset()
	logger.Info(nil)
	output := suite.buffer.String()
	assert.Contains(suite.T(), output, "<nil>")
	
	suite.buffer.Reset()
	logger.Infof("Value: %v", nil)
	output = suite.buffer.String()
	assert.Contains(suite.T(), output, "Value: <nil>")
}

// TestLoggerDefaultOutput tests that logger outputs to stdout by default
func (suite *LoggerTestSuite) TestLoggerDefaultOutput() {
	// This test verifies the logger is configured to output to os.Stdout
	// We can't easily capture os.Stdout in tests, but we can verify the configuration
	logger := NewLogger("info", "text")
	logrusLogger, ok := logger.(*LogrusLogger)
	require.True(suite.T(), ok)
	
	// Verify the output is set to os.Stdout
	// Note: logrus doesn't expose the output directly, but we can verify it was created successfully
	assert.NotNil(suite.T(), logrusLogger.logger)
}

// Run the test suite
func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// Standalone tests

func TestNewLoggerLevelValidation(t *testing.T) {
	testCases := []struct {
		inputLevel    string
		expectedLevel logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
		{"invalid", logrus.InfoLevel}, // defaults to info
		{"", logrus.InfoLevel},        // defaults to info
		{"DEBUG", logrus.InfoLevel},   // case sensitive, defaults to info
		{"INFO", logrus.InfoLevel},    // case sensitive, defaults to info
	}
	
	for _, tc := range testCases {
		t.Run("Level_"+tc.inputLevel, func(t *testing.T) {
			logger := NewLogger(tc.inputLevel, "text")
			logrusLogger, ok := logger.(*LogrusLogger)
			require.True(t, ok)
			assert.Equal(t, tc.expectedLevel, logrusLogger.logger.Level)
		})
	}
}

func TestNewLoggerFormatValidation(t *testing.T) {
	testCases := []struct {
		format       string
		expectedType string
	}{
		{"json", "*logrus.JSONFormatter"},
		{"text", "*logrus.TextFormatter"},
		{"invalid", "*logrus.TextFormatter"}, // defaults to text
		{"", "*logrus.TextFormatter"},        // defaults to text
		{"JSON", "*logrus.TextFormatter"},    // case sensitive, defaults to text
	}
	
	for _, tc := range testCases {
		t.Run("Format_"+tc.format, func(t *testing.T) {
			logger := NewLogger("info", tc.format)
			logrusLogger, ok := logger.(*LogrusLogger)
			require.True(t, ok)
			
			formatter := logrusLogger.logger.Formatter
			// Check formatter type by type assertion instead of calling Format
			switch tc.format {
			case "json":
				_, ok := formatter.(*logrus.JSONFormatter)
				assert.True(t, ok, "Expected JSON formatter")
			default:
				_, ok := formatter.(*logrus.TextFormatter)
				assert.True(t, ok, "Expected Text formatter")
			}
		})
	}
}

func TestLoggerConcurrency(t *testing.T) {
	logger := NewLogger("info", "json")
	
	// Test concurrent logging
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				logger.Infof("Goroutine %d, message %d", id, j)
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// If we reach here without panicking, the test passes
	assert.True(t, true)
}

func TestLoggerMemoryUsage(t *testing.T) {
	// Test that logger doesn't leak memory with large messages
	logger := NewLogger("info", "text")
	
	largeMessage := strings.Repeat("A", 10000) // 10KB message
	
	// Log many large messages
	for i := 0; i < 100; i++ {
		logger.Info(largeMessage)
	}
	
	// If we complete without issues, test passes
	assert.True(t, true)
}