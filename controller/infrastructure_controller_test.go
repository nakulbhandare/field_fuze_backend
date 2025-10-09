package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fieldfuze-backend/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockInfrastructureService implements InfrastructureServiceInterface for testing
type MockInfrastructureService struct {
	mock.Mock
}

func (m *MockInfrastructureService) GetWorkerStatus(ctx context.Context) (*models.ExecutionResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ExecutionResult), args.Error(1)
}

func (m *MockInfrastructureService) RestartWorker(ctx context.Context, force bool) (*models.ServiceRestartResult, error) {
	args := m.Called(ctx, force)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceRestartResult), args.Error(1)
}

func (m *MockInfrastructureService) IsWorkerHealthy() (bool, string, error) {
	args := m.Called()
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *MockInfrastructureService) AutoRestartIfNeeded(ctx context.Context) (*models.ServiceRestartResult, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceRestartResult), args.Error(1)
}

// InfrastructureControllerTestSuite contains the test suite for InfrastructureController
type InfrastructureControllerTestSuite struct {
	suite.Suite
	infraController *InfrastructureController
	mockService     *MockInfrastructureService
	mockLogger      *MockControllerLogger
	ctx             context.Context
	router          *gin.Engine
}

func (suite *InfrastructureControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	suite.mockService = &MockInfrastructureService{}
	suite.mockLogger = &MockControllerLogger{}
	
	// Set up comprehensive mock expectations for all logger patterns
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything).Maybe()
	suite.mockLogger.On("Error", mock.Anything).Maybe()
	suite.mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Infof", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Errorf", mock.Anything, mock.Anything).Maybe()
	
	suite.infraController = NewInfrastructureController(suite.ctx, suite.mockService, suite.mockLogger)
	suite.router = gin.New()
}

func TestInfrastructureControllerTestSuite(t *testing.T) {
	suite.Run(t, new(InfrastructureControllerTestSuite))
}

// TestNewInfrastructureController tests the constructor
func (suite *InfrastructureControllerTestSuite) TestNewInfrastructureController() {
	controller := NewInfrastructureController(suite.ctx, suite.mockService, suite.mockLogger)
	
	assert.NotNil(suite.T(), controller)
	assert.Equal(suite.T(), suite.ctx, controller.ctx)
	assert.Equal(suite.T(), suite.mockService, controller.service)
	assert.Equal(suite.T(), suite.mockLogger, controller.logger)
}

// TestGetWorkerStatus tests successful worker status retrieval
func (suite *InfrastructureControllerTestSuite) TestGetWorkerStatus() {
	expectedResult := &models.ExecutionResult{
		Status:  models.StatusCompleted,
		Success: true,
	}
	
	suite.mockService.On("GetWorkerStatus", suite.ctx).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/status", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/status", suite.infraController.GetWorkerStatus)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Infrastructure is ready and healthy", response.Message)
}

// TestGetWorkerStatusFailed tests failed worker status
func (suite *InfrastructureControllerTestSuite) TestGetWorkerStatusFailed() {
	expectedResult := &models.ExecutionResult{
		Status:  models.StatusFailed,
		Success: false,
	}
	
	suite.mockService.On("GetWorkerStatus", suite.ctx).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/status", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/status", suite.infraController.GetWorkerStatus)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusServiceUnavailable, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestGetWorkerStatusInProgress tests in-progress worker status
func (suite *InfrastructureControllerTestSuite) TestGetWorkerStatusInProgress() {
	expectedResult := &models.ExecutionResult{
		Status:  models.StatusRunning,
		Success: false,
	}
	
	suite.mockService.On("GetWorkerStatus", suite.ctx).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/status", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/status", suite.infraController.GetWorkerStatus)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusAccepted, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "in_progress", response.Status)
	assert.Equal(suite.T(), "Infrastructure setup is running", response.Message)
}

// TestGetWorkerStatusServiceError tests service error during status retrieval
func (suite *InfrastructureControllerTestSuite) TestGetWorkerStatusServiceError() {
	suite.mockService.On("GetWorkerStatus", suite.ctx).Return(nil, errors.New("service error"))
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/status", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/status", suite.infraController.GetWorkerStatus)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to retrieve worker status", response.Message)
}

// TestRestartWorker tests successful worker restart
func (suite *InfrastructureControllerTestSuite) TestRestartWorker() {
	expectedResult := &models.ServiceRestartResult{
		Status:  "completed",
		Output:  "Worker restarted successfully",
	}
	
	suite.mockService.On("RestartWorker", suite.ctx, false).Return(expectedResult, nil)
	
	requestBody := models.WorkerRestartRequest{Force: false}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/restart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/restart", suite.infraController.RestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Worker restart initiated successfully", response.Message)
}

// TestRestartWorkerWithForce tests forced worker restart
func (suite *InfrastructureControllerTestSuite) TestRestartWorkerWithForce() {
	expectedResult := &models.ServiceRestartResult{
		Status:  "completed",
		Output:  "Worker force restarted successfully",
	}
	
	suite.mockService.On("RestartWorker", suite.ctx, true).Return(expectedResult, nil)
	
	requestBody := models.WorkerRestartRequest{Force: true}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/restart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/restart", suite.infraController.RestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Worker restart initiated successfully", response.Message)
}

// TestRestartWorkerNoBody tests restart worker with no request body
func (suite *InfrastructureControllerTestSuite) TestRestartWorkerNoBody() {
	expectedResult := &models.ServiceRestartResult{
		Status:  "completed",
		Output:  "Worker restarted successfully",
	}
	
	suite.mockService.On("RestartWorker", suite.ctx, false).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/restart", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/restart", suite.infraController.RestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestRestartWorkerConflict tests restart worker when worker is running
func (suite *InfrastructureControllerTestSuite) TestRestartWorkerConflict() {
	suite.mockService.On("RestartWorker", suite.ctx, false).Return(nil, errors.New("worker is running"))
	
	requestBody := models.WorkerRestartRequest{Force: false}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/restart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/restart", suite.infraController.RestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Worker is currently running", response.Message)
}

// TestRestartWorkerServiceError tests restart worker service error
func (suite *InfrastructureControllerTestSuite) TestRestartWorkerServiceError() {
	suite.mockService.On("RestartWorker", suite.ctx, false).Return(nil, errors.New("service error"))
	
	requestBody := models.WorkerRestartRequest{Force: false}
	body, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/restart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/restart", suite.infraController.RestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to restart worker", response.Message)
}

// TestCheckWorkerHealth tests successful health check
func (suite *InfrastructureControllerTestSuite) TestCheckWorkerHealth() {
	suite.mockService.On("IsWorkerHealthy").Return(true, "Worker is running normally", nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/health", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/health", suite.infraController.CheckWorkerHealth)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Worker health check completed", response.Message)
	
	// Check health data
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), true, data["healthy"])
	assert.Equal(suite.T(), "healthy", data["status"])
	assert.Equal(suite.T(), "Worker is running normally", data["reason"])
}

// TestCheckWorkerHealthUnhealthy tests unhealthy worker health check
func (suite *InfrastructureControllerTestSuite) TestCheckWorkerHealthUnhealthy() {
	suite.mockService.On("IsWorkerHealthy").Return(false, "Worker is not responding", nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/health", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/health", suite.infraController.CheckWorkerHealth)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	
	// Check health data
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), false, data["healthy"])
	assert.Equal(suite.T(), "unhealthy", data["status"])
	assert.Equal(suite.T(), "Worker is not responding", data["reason"])
}

// TestCheckWorkerHealthServiceError tests health check service error
func (suite *InfrastructureControllerTestSuite) TestCheckWorkerHealthServiceError() {
	suite.mockService.On("IsWorkerHealthy").Return(false, "", errors.New("health check failed"))
	
	req, _ := http.NewRequest(http.MethodGet, "/infrastructure/worker/health", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/infrastructure/worker/health", suite.infraController.CheckWorkerHealth)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to check worker health", response.Message)
}

// TestAutoRestartWorker tests successful auto-restart (not needed)
func (suite *InfrastructureControllerTestSuite) TestAutoRestartWorkerNotNeeded() {
	expectedResult := &models.ServiceRestartResult{
		Status:  "not_needed",
		Output:  "Worker is healthy",
	}
	
	suite.mockService.On("AutoRestartIfNeeded", suite.ctx).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/auto-restart", nil)
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/auto-restart", suite.infraController.AutoRestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Worker is healthy, no restart needed", response.Message)
}

// TestAutoRestartWorkerCompleted tests auto-restart completed
func (suite *InfrastructureControllerTestSuite) TestAutoRestartWorkerCompleted() {
	expectedResult := &models.ServiceRestartResult{
		Status:  "completed",
		Output:  "Worker was restarted",
	}
	
	suite.mockService.On("AutoRestartIfNeeded", suite.ctx).Return(expectedResult, nil)
	
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/auto-restart", nil)
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/auto-restart", suite.infraController.AutoRestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Worker was unhealthy and has been restarted", response.Message)
}

// TestAutoRestartWorkerServiceError tests auto-restart service error
func (suite *InfrastructureControllerTestSuite) TestAutoRestartWorkerServiceError() {
	suite.mockService.On("AutoRestartIfNeeded", suite.ctx).Return(nil, errors.New("auto-restart failed"))
	
	req, _ := http.NewRequest(http.MethodPost, "/infrastructure/worker/auto-restart", nil)
	w := httptest.NewRecorder()
	
	suite.router.POST("/infrastructure/worker/auto-restart", suite.infraController.AutoRestartWorker)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to auto-restart worker", response.Message)
}

// TestMapWorkerStatusToHTTP tests various status mappings
func (suite *InfrastructureControllerTestSuite) TestMapWorkerStatusToHTTP() {
	testCases := []struct {
		name           string
		executionResult *models.ExecutionResult
		expectedCode   int
		expectedStatus string
	}{
		{
			name: "Completed Successfully",
			executionResult: &models.ExecutionResult{Status: models.StatusCompleted, Success: true},
			expectedCode:   http.StatusOK,
			expectedStatus: "success",
		},
		{
			name: "Completed with Issues",
			executionResult: &models.ExecutionResult{Status: models.StatusCompleted, Success: false},
			expectedCode:   http.StatusOK,
			expectedStatus: "warning",
		},
		{
			name: "Failed",
			executionResult: &models.ExecutionResult{Status: models.StatusFailed, Success: false},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: "error",
		},
		{
			name: "Running",
			executionResult: &models.ExecutionResult{Status: models.StatusRunning, Success: false},
			expectedCode:   http.StatusAccepted,
			expectedStatus: "in_progress",
		},
		{
			name: "Retrying",
			executionResult: &models.ExecutionResult{Status: models.StatusRetrying, Success: false},
			expectedCode:   http.StatusAccepted,
			expectedStatus: "retrying",
		},
		{
			name: "Deleting",
			executionResult: &models.ExecutionResult{Status: models.StatusDeleting, Success: false},
			expectedCode:   http.StatusAccepted,
			expectedStatus: "deleting",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			code, status := suite.infraController.mapWorkerStatusToHTTP(tc.executionResult)
			assert.Equal(suite.T(), tc.expectedCode, code)
			assert.Equal(suite.T(), tc.expectedStatus, status)
		})
	}
}

// TestGetStatusMessage tests various status messages
func (suite *InfrastructureControllerTestSuite) TestGetStatusMessage() {
	testCases := []struct {
		name           string
		executionResult *models.ExecutionResult
		expectedMessage string
	}{
		{
			name: "Completed Successfully",
			executionResult: &models.ExecutionResult{Status: models.StatusCompleted, Success: true},
			expectedMessage: "Infrastructure is ready and healthy",
		},
		{
			name: "Completed with Warnings",
			executionResult: &models.ExecutionResult{Status: models.StatusCompleted, Success: false},
			expectedMessage: "Infrastructure setup completed with warnings",
		},
		{
			name: "Failed",
			executionResult: &models.ExecutionResult{Status: models.StatusFailed, Success: false},
			expectedMessage: "Infrastructure setup failed - manual intervention may be required",
		},
		{
			name: "Running",
			executionResult: &models.ExecutionResult{Status: models.StatusRunning, Success: false},
			expectedMessage: "Infrastructure setup is running",
		},
		{
			name: "Creating Tables",
			executionResult: &models.ExecutionResult{Status: models.StatusCreatingTables, Success: false},
			expectedMessage: "Creating DynamoDB tables",
		},
		{
			name: "Retrying with Count",
			executionResult: &models.ExecutionResult{
				Status: models.StatusRetrying,
				Metadata: map[string]interface{}{
					"retry_count": 2,
				},
			},
			expectedMessage: "Retrying infrastructure setup (attempt 3)",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			message := suite.infraController.getStatusMessage(tc.executionResult)
			assert.Equal(suite.T(), tc.expectedMessage, message)
		})
	}
}