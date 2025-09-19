package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the interface for logging
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

// LogrusLogger wraps logrus.Logger to implement our Logger interface
type LogrusLogger struct {
	logger *logrus.Logger
}

// NewLogger creates a new logger instance
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

// Debug logs a debug message
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Info logs an info message
func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Infof logs a formatted info message
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warn logs a warning message
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Error logs an error message
func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Errorf logs a formatted error message
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
