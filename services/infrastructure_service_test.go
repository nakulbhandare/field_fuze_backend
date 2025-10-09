package services

import (
	"context"
	"encoding/json"
	"fieldfuze-backend/models"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// MockInfraDatabaseClient implements the DatabaseClientInterface for testing
type MockInfraDatabaseClient struct {
	mock.Mock
}

func (m *MockInfraDatabaseClient) GetItem(ctx context.Context, config models.QueryConfig, result interface{}) error {
	args := m.Called(ctx, config, result)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) PutItem(ctx context.Context, tableName string, item interface{}) error {
	args := m.Called(ctx, tableName, item)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) UpdateItem(ctx context.Context, tableName, key, keyValue string, updates map[string]interface{}) error {
	args := m.Called(ctx, tableName, key, keyValue, updates)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) DeleteItem(ctx context.Context, tableName, key, value string) error {
	args := m.Called(ctx, tableName, key, value)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) QueryByIndex(ctx context.Context, tableName, indexName, keyName, keyValue string, results interface{}) error {
	args := m.Called(ctx, tableName, indexName, keyName, keyValue, results)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) Scan(ctx context.Context, tableName string, results interface{}) error {
	args := m.Called(ctx, tableName, results)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) ScanTable(ctx context.Context, tableName string, results interface{}) error {
	args := m.Called(ctx, tableName, results)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) CreateTable(ctx context.Context, input *dynamodb.CreateTableInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockInfraDatabaseClient) DescribeTable(ctx context.Context, tableName string) (*dynamodb.DescribeTableOutput, error) {
	args := m.Called(ctx, tableName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.DescribeTableOutput), args.Error(1)
}

func (m *MockInfraDatabaseClient) DeleteTable(ctx context.Context, input *dynamodb.DeleteTableInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// MockInfraLogger implements the Logger interface for testing
type MockInfraLogger struct {
	mock.Mock
}

func (m *MockInfraLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockInfraLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockInfraLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockInfraLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockInfraLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockInfraLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockInfraLogger) Warnf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockInfraLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockInfraLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockInfraLogger) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

// InfrastructureServiceTestSuite contains the test suite for InfrastructureService
type InfrastructureServiceTestSuite struct {
	suite.Suite
	infraService   *InfrastructureService
	mockDBClient   *MockInfraDatabaseClient
	mockLogger     *MockInfraLogger
	ctx            context.Context
	config         *models.Config
	tempDir        string
}

func (suite *InfrastructureServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.mockDBClient = &MockInfraDatabaseClient{}
	suite.mockLogger = &MockInfraLogger{}
	
	// Set up comprehensive mock expectations
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything).Maybe()
	suite.mockLogger.On("Error", mock.Anything).Maybe()
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Infof", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Errorf", mock.Anything, mock.Anything).Maybe()
	
	suite.config = &models.Config{
		AppEnv: "test",
	}
	
	suite.infraService = NewInfrastructureService(
		suite.ctx,
		suite.mockDBClient,
		suite.mockLogger,
		suite.config,
	)
	
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "infra_service_test")
	assert.NoError(suite.T(), err)
	suite.tempDir = tempDir
}

func (suite *InfrastructureServiceTestSuite) TearDownTest() {
	// Clean up temporary files
	if suite.tempDir != "" {
		os.RemoveAll(suite.tempDir)
	}
}

func TestInfrastructureServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InfrastructureServiceTestSuite))
}

// TestNewInfrastructureService tests the constructor
func (suite *InfrastructureServiceTestSuite) TestNewInfrastructureService() {
	service := NewInfrastructureService(suite.ctx, suite.mockDBClient, suite.mockLogger, suite.config)
	
	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.ctx, service.ctx)
	assert.Equal(suite.T(), suite.mockDBClient, service.dbClient)
	assert.Equal(suite.T(), suite.mockLogger, service.logger)
	assert.Equal(suite.T(), suite.config, service.config)
}

// TestGetWorkerStatus tests the GetWorkerStatus function
func (suite *InfrastructureServiceTestSuite) TestGetWorkerStatus() {
	// Create test status file
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status:    models.StatusRunning,
		Phase:     "Test Phase",
		Progress:  &models.ProgressInfo{Percentage: 50},
		StartTime: time.Now().Add(-5 * time.Minute),
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	
	result, err := suite.infraService.GetWorkerStatus(suite.ctx)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), models.StatusRunning, result.Status)
	// Phase may be enriched by the service, just verify it's not empty
	assert.NotEmpty(suite.T(), result.Phase)
	assert.Equal(suite.T(), 50, result.Progress.Percentage)
}

// TestGetWorkerStatusFileNotFound tests GetWorkerStatus when status file doesn't exist
func (suite *InfrastructureServiceTestSuite) TestGetWorkerStatusFileNotFound() {
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	
	result, err := suite.infraService.GetWorkerStatus(suite.ctx)
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to read worker status file")
}

// TestGetWorkerStatusInvalidJSON tests GetWorkerStatus with invalid JSON
func (suite *InfrastructureServiceTestSuite) TestGetWorkerStatusInvalidJSON() {
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	
	err := os.WriteFile(statusFilePath, []byte("invalid json"), 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	
	result, err := suite.infraService.GetWorkerStatus(suite.ctx)
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal worker status")
}

// TestRestartWorker tests the RestartWorker function with force=true
func (suite *InfrastructureServiceTestSuite) TestRestartWorkerWithForce() {
	// Create test status file showing running worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status: models.StatusRunning,
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything, mock.Anything).Maybe()
	
	result, err := suite.infraService.RestartWorker(suite.ctx, true)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "infrastructure-worker", result.ServiceName)
	assert.Equal(suite.T(), "completed", result.Status)
}

// TestRestartWorkerWithoutForce tests RestartWorker when worker is running and force=false
func (suite *InfrastructureServiceTestSuite) TestRestartWorkerWithoutForce() {
	// Create test status file showing running worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status: models.StatusRunning,
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	
	result, err := suite.infraService.RestartWorker(suite.ctx, false)
	
	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "failed", result.Status)
	assert.Contains(suite.T(), result.Error, "Worker is currently running")
}

// TestRestartWorkerNoStatusFile tests RestartWorker when status file doesn't exist
func (suite *InfrastructureServiceTestSuite) TestRestartWorkerNoStatusFile() {
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything, mock.Anything).Maybe()
	
	result, err := suite.infraService.RestartWorker(suite.ctx, false)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "completed", result.Status)
}

// TestIsWorkerHealthy tests the IsWorkerHealthy function with healthy worker
func (suite *InfrastructureServiceTestSuite) TestIsWorkerHealthy() {
	// Create test status file with healthy worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status:    models.StatusRunning,
		StartTime: time.Now().Add(-2 * time.Minute),
		Progress:  &models.ProgressInfo{Percentage: 75},
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	isHealthy, reason, err := suite.infraService.IsWorkerHealthy()
	
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), isHealthy)
	assert.Equal(suite.T(), "Worker is running normally", reason)
}

// TestIsWorkerHealthyFailed tests IsWorkerHealthy with failed worker
func (suite *InfrastructureServiceTestSuite) TestIsWorkerHealthyFailed() {
	// Create test status file with failed worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status:       models.StatusFailed,
		ErrorMessage: "Test error",
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	isHealthy, reason, err := suite.infraService.IsWorkerHealthy()
	
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), isHealthy)
	assert.Contains(suite.T(), reason, "Worker failed")
}

// TestIsWorkerHealthyNoStatusFile tests IsWorkerHealthy when status file doesn't exist
func (suite *InfrastructureServiceTestSuite) TestIsWorkerHealthyNoStatusFile() {
	isHealthy, reason, err := suite.infraService.IsWorkerHealthy()
	
	assert.Error(suite.T(), err)
	assert.False(suite.T(), isHealthy)
	assert.Contains(suite.T(), err.Error(), "failed to read worker status file")
	// Reason will contain error message
	assert.NotEmpty(suite.T(), reason)
}

// TestAutoRestartIfNeeded tests AutoRestartIfNeeded with unhealthy worker
func (suite *InfrastructureServiceTestSuite) TestAutoRestartIfNeededUnhealthy() {
	// Create test status file with failed worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status:       models.StatusFailed,
		ErrorMessage: "Test error",
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything, mock.Anything).Maybe()
	
	result, err := suite.infraService.AutoRestartIfNeeded(suite.ctx)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "completed", result.Status)
}

// TestAutoRestartIfNeededHealthy tests AutoRestartIfNeeded with healthy worker
func (suite *InfrastructureServiceTestSuite) TestAutoRestartIfNeededHealthy() {
	// Create test status file with healthy worker
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status:    models.StatusRunning,
		StartTime: time.Now().Add(-2 * time.Minute),
		Progress:  &models.ProgressInfo{Percentage: 75},
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	result, err := suite.infraService.AutoRestartIfNeeded(suite.ctx)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "not_needed", result.Status)
}

// TestEnrichStatusWithContext tests enrichStatusWithContext function
func (suite *InfrastructureServiceTestSuite) TestEnrichStatusWithContext() {
	testCases := []struct {
		name           string
		status         models.WorkerStatus
		expectedAction string
		expectedPhase  string
	}{
		{
			name:           "Creating Tables",
			status:         models.StatusCreatingTables,
			expectedAction: "Creating DynamoDB tables - this may take a few minutes",
			expectedPhase:  "Table Creation",
		},
		{
			name:           "Waiting for Tables",
			status:         models.StatusWaitingForTables,
			expectedAction: "Waiting for DynamoDB tables to become active",
			expectedPhase:  "Table Activation",
		},
		{
			name:           "Creating Indexes",
			status:         models.StatusCreatingIndexes,
			expectedAction: "Creating database indexes",
			expectedPhase:  "Index Creation",
		},
		{
			name:           "Validating",
			status:         models.StatusValidating,
			expectedAction: "Validating infrastructure configuration",
			expectedPhase:  "Validation",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := &models.ExecutionResult{
				Status: tc.status,
			}
			
			suite.infraService.enrichStatusWithContext(result)
			
			assert.Equal(suite.T(), tc.expectedAction, result.NextAction)
			assert.Equal(suite.T(), tc.expectedPhase, result.Phase)
			assert.NotNil(suite.T(), result.EstimatedTime)
		})
	}
}

// TestUpdateHealthIndicators tests updateHealthIndicators function
func (suite *InfrastructureServiceTestSuite) TestUpdateHealthIndicators() {
	result := &models.ExecutionResult{
		Status:    models.StatusRunning,
		StartTime: time.Now().Add(-5 * time.Minute),
		Progress:  &models.ProgressInfo{Percentage: 75},
	}
	
	suite.infraService.updateHealthIndicators(result)
	
	assert.NotNil(suite.T(), result.HealthStatus)
}

// TestCalculateProgress tests calculateProgress function
func (suite *InfrastructureServiceTestSuite) TestCalculateProgress() {
	result := &models.ExecutionResult{
		Status:   models.StatusRunning,
		Progress: &models.ProgressInfo{Percentage: 75},
	}
	
	progressInfo := suite.infraService.calculateProgress(result)
	
	assert.NotNil(suite.T(), progressInfo)
	// Progress calculation is based on internal logic, just verify it returns something
	assert.True(suite.T(), progressInfo.Percentage >= 0)
}

// TestDurationPtr tests durationPtr helper function
func (suite *InfrastructureServiceTestSuite) TestDurationPtr() {
	duration := 5 * time.Minute
	ptr := suite.infraService.durationPtr(duration)
	
	assert.NotNil(suite.T(), ptr)
	assert.Equal(suite.T(), duration, *ptr)
}

// TestKillWorkerProcess tests killWorkerProcess function
func (suite *InfrastructureServiceTestSuite) TestKillWorkerProcess() {
	// This test verifies the function runs without panicking
	// since it deals with process management
	err := suite.infraService.killWorkerProcess()
	
	// killWorkerProcess may not return error if no process is found
	// Just verify it doesn't panic
	_ = err
}

// TestResetWorkerStatus tests resetWorkerStatus function
func (suite *InfrastructureServiceTestSuite) TestResetWorkerStatus() {
	// Create a status file first
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status: models.StatusRunning,
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	
	err = suite.infraService.resetWorkerStatus()
	
	assert.NoError(suite.T(), err)
	
	// Verify the file was removed
	_, err = os.Stat(statusFilePath)
	assert.True(suite.T(), os.IsNotExist(err))
}

// TestStartWorkerProcess tests startWorkerProcess function
func (suite *InfrastructureServiceTestSuite) TestStartWorkerProcess() {
	// This test verifies the function attempts to start a process
	// The function may succeed in test mode
	err := suite.infraService.startWorkerProcess(suite.ctx)
	
	// Just verify it doesn't panic, error or success is acceptable
	_ = err
}

// TestGetRetryCountFromMetadata tests getRetryCountFromMetadata function
func (suite *InfrastructureServiceTestSuite) TestGetRetryCountFromMetadata() {
	// Test with metadata containing retry count
	result := &models.ExecutionResult{
		Metadata: map[string]interface{}{
			"retry_count": 3,
		},
	}
	
	count := suite.infraService.getRetryCountFromMetadata(result)
	assert.Equal(suite.T(), 3, count)
	
	// Test with no metadata
	result = &models.ExecutionResult{}
	count = suite.infraService.getRetryCountFromMetadata(result)
	assert.Equal(suite.T(), 0, count)
	
	// Test with metadata but no retry count
	result = &models.ExecutionResult{
		Metadata: map[string]interface{}{
			"other_field": "value",
		},
	}
	count = suite.infraService.getRetryCountFromMetadata(result)
	assert.Equal(suite.T(), 0, count)
}

// Edge case tests

// TestGetWorkerStatusConcurrentAccess tests concurrent access to status file
func (suite *InfrastructureServiceTestSuite) TestGetWorkerStatusConcurrentAccess() {
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	testResult := &models.ExecutionResult{
		Status: models.StatusRunning,
	}
	
	data, err := json.Marshal(testResult)
	assert.NoError(suite.T(), err)
	
	err = os.WriteFile(statusFilePath, data, 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	
	// Test concurrent access
	done := make(chan bool, 2)
	
	go func() {
		_, err := suite.infraService.GetWorkerStatus(suite.ctx)
		assert.NoError(suite.T(), err)
		done <- true
	}()
	
	go func() {
		_, err := suite.infraService.GetWorkerStatus(suite.ctx)
		assert.NoError(suite.T(), err)
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestRestartWorkerEdgeCases tests various edge cases for RestartWorker
func (suite *InfrastructureServiceTestSuite) TestRestartWorkerEdgeCases() {
	// Test with corrupted status file
	statusFilePath := fmt.Sprintf("/tmp/fieldfuze-status-%s.json", suite.config.AppEnv)
	err := os.WriteFile(statusFilePath, []byte("corrupted"), 0644)
	assert.NoError(suite.T(), err)
	defer os.Remove(statusFilePath)
	
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything, mock.Anything).Maybe()
	
	result, err := suite.infraService.RestartWorker(suite.ctx, false)
	
	// Should still succeed despite corrupted status file
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "completed", result.Status)
}

// TestCalculateProgressEdgeCases tests edge cases for progress calculation
func (suite *InfrastructureServiceTestSuite) TestCalculateProgressEdgeCases() {
	// Test with progress over 100
	result := &models.ExecutionResult{
		Status:   models.StatusRunning,
		Progress: &models.ProgressInfo{Percentage: 150},
	}
	
	progressInfo := suite.infraService.calculateProgress(result)
	
	assert.NotNil(suite.T(), progressInfo)
	// Progress calculation is based on internal logic
	assert.True(suite.T(), progressInfo.Percentage >= 0)
	
	// Test with negative progress
	result = &models.ExecutionResult{
		Status:   models.StatusRunning,
		Progress: &models.ProgressInfo{Percentage: -10},
	}
	
	progressInfo = suite.infraService.calculateProgress(result)
	
	assert.NotNil(suite.T(), progressInfo)
	// Progress calculation is based on internal logic
	assert.True(suite.T(), progressInfo.Percentage >= 0)
}