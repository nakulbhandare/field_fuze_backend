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

// MockOrganizationService implements OrganizationServiceInterface for testing
type MockOrganizationService struct {
	mock.Mock
}

func (m *MockOrganizationService) CreateOrganization(ctx context.Context, org *models.Organization, createdBy string) (*models.Organization, error) {
	args := m.Called(ctx, org, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizations(status string) ([]*models.Organization, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) GetOrganizationByID(id string) (*models.Organization, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) UpdateOrganization(id string, org *models.Organization, updatedBy string) (*models.Organization, error) {
	args := m.Called(id, org, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) DeleteOrganization(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrganizationService) DeleteOrganizationAssignment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrganizationService) GetOrganizationAssignmentsByStatus(status string) ([]*models.Organization, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) UpdateOrganizationAssignment(id string, org *models.Organization, updatedBy string) (*models.Organization, error) {
	args := m.Called(id, org, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationService) ValidateOrganizationEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockOrganizationService) ValidateOrganizationPhone(phone string) error {
	args := m.Called(phone)
	return args.Error(0)
}

// OrganizationControllerTestSuite contains the test suite for OrganizationController
type OrganizationControllerTestSuite struct {
	suite.Suite
	orgController *OrganizationController
	mockService   *MockOrganizationService
	mockLogger    *MockControllerLogger
	ctx           context.Context
	router        *gin.Engine
}

func (suite *OrganizationControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	suite.mockService = &MockOrganizationService{}
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
	
	suite.orgController = NewOrganizationController(suite.ctx, suite.mockService, suite.mockLogger)
	suite.router = gin.New()
}

func TestOrganizationControllerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationControllerTestSuite))
}

// TestNewOrganizationController tests the constructor
func (suite *OrganizationControllerTestSuite) TestNewOrganizationController() {
	controller := NewOrganizationController(suite.ctx, suite.mockService, suite.mockLogger)
	
	assert.NotNil(suite.T(), controller)
	assert.Equal(suite.T(), suite.ctx, controller.ctx)
	assert.Equal(suite.T(), suite.mockService, controller.organizationService)
	assert.Equal(suite.T(), suite.mockLogger, controller.logger)
	assert.NotNil(suite.T(), controller.validator)
}

// TestCreateOrganization tests successful organization creation
func (suite *OrganizationControllerTestSuite) TestCreateOrganization() {
	orgReq := models.Organization{
		Name:    "Test Organization",
		Email:   "test@example.com",
		Phone:   "+1234567890",
		Status:  "active",
	}
	
	expectedOrg := &models.Organization{
		ID:      "org-123",
		Name:    "Test Organization",
		Email:   "test@example.com",
		Phone:   "+1234567890",
		Status:  "active",
	}
	
	suite.mockService.On("CreateOrganization", suite.ctx, mock.MatchedBy(func(org *models.Organization) bool {
		return org.Name == "Test Organization" && org.Email == "test@example.com"
	}), "user-123").Return(expectedOrg, nil)
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "user-123"})
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Organization created successfully", response.Message)
}

// TestCreateOrganizationInvalidJSON tests invalid JSON binding
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationInvalidJSON() {
	invalidJSON := `{"name": "Test Org", "email": invalid-email}`
	
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "user-123"})
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Invalid request", response.Message)
}

// TestCreateOrganizationValidationFailure tests validation failure
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationValidationFailure() {
	orgReq := models.Organization{
		Name:  "", // Empty name should fail validation
		Email: "invalid-email",
	}
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "user-123"})
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Validation failed", response.Message)
}

// TestCreateOrganizationNoJWTClaims tests missing JWT claims
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationNoJWTClaims() {
	orgReq := models.Organization{
		Name:    "Test Organization",
		Email:   "test@example.com",
		Status:  "active",
	}
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", suite.orgController.CreateOrganization)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Authentication required", response.Message)
}

// TestCreateOrganizationInvalidJWTClaims tests invalid JWT claims type
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationInvalidJWTClaims() {
	orgReq := models.Organization{
		Name:    "Test Organization",
		Email:   "test@example.com",
		Status:  "active",
	}
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", "invalid-claims") // Invalid type
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Invalid token claims", response.Message)
}

// TestCreateOrganizationServiceError tests service error during creation
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationServiceError() {
	orgReq := models.Organization{
		Name:    "Test Organization",
		Email:   "test@example.com",
		Status:  "active",
	}
	
	suite.mockService.On("CreateOrganization", suite.ctx, mock.Anything, "user-123").Return(nil, errors.New("database error"))
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "user-123"})
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to create role", response.Message)
}

// TestCreateOrganizationConflict tests organization already exists error
func (suite *OrganizationControllerTestSuite) TestCreateOrganizationConflict() {
	orgReq := models.Organization{
		Name:    "Test Organization",
		Email:   "test@example.com",
		Status:  "active",
	}
	
	suite.mockService.On("CreateOrganization", suite.ctx, mock.Anything, "user-123").Return(nil, errors.New("organization with this name already exists"))
	
	body, _ := json.Marshal(orgReq)
	req, _ := http.NewRequest(http.MethodPost, "/organization", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.POST("/organization", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "user-123"})
		suite.orgController.CreateOrganization(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to create role", response.Message)
}

// TestGetOrganizations tests successful organization retrieval
func (suite *OrganizationControllerTestSuite) TestGetOrganizations() {
	expectedOrgs := []*models.Organization{
		{
			ID:     "org-1",
			Name:   "Organization 1",
			Email:  "org1@example.com",
			Status: "active",
		},
		{
			ID:     "org-2",
			Name:   "Organization 2",
			Email:  "org2@example.com",
			Status: "active",
		},
	}
	
	suite.mockService.On("GetOrganizations", "").Return(expectedOrgs, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/organization", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/organization", suite.orgController.GetOrganizations)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Roles retrieved successfully", response.Message)
	
	// Check pagination data structure
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Contains(suite.T(), data, "organizations")
	assert.Contains(suite.T(), data, "pagination")
}

// TestGetOrganizationsWithPagination tests pagination parameters
func (suite *OrganizationControllerTestSuite) TestGetOrganizationsWithPagination() {
	expectedOrgs := []*models.Organization{
		{ID: "org-1", Name: "Organization 1", Status: "active"},
		{ID: "org-2", Name: "Organization 2", Status: "active"},
		{ID: "org-3", Name: "Organization 3", Status: "active"},
	}
	
	suite.mockService.On("GetOrganizations", "").Return(expectedOrgs, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/organization?page=2&limit=2", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/organization", suite.orgController.GetOrganizations)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	
	// Check pagination values
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	
	pagination, ok := data["pagination"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(2), pagination["page"])
	assert.Equal(suite.T(), float64(2), pagination["limit"])
	assert.Equal(suite.T(), float64(3), pagination["total"])
}

// TestGetOrganizationsServiceError tests service error during retrieval
func (suite *OrganizationControllerTestSuite) TestGetOrganizationsServiceError() {
	suite.mockService.On("GetOrganizations", "").Return(nil, errors.New("database error"))
	
	req, _ := http.NewRequest(http.MethodGet, "/organization", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/organization", suite.orgController.GetOrganizations)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
	assert.Equal(suite.T(), "Failed to get roles", response.Message)
}

// TestGetOrganizationsInvalidPageParam tests invalid page parameter
func (suite *OrganizationControllerTestSuite) TestGetOrganizationsInvalidPageParam() {
	expectedOrgs := []*models.Organization{
		{ID: "org-1", Name: "Organization 1", Status: "active"},
	}
	
	suite.mockService.On("GetOrganizations", "").Return(expectedOrgs, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/organization?page=invalid&limit=abc", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/organization", suite.orgController.GetOrganizations)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	
	// Should default to page=1, limit=10
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	
	pagination, ok := data["pagination"].(map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), float64(1), pagination["page"])
	assert.Equal(suite.T(), float64(10), pagination["limit"])
}

// TestGetOrganizationsEmptyResult tests empty organization list
func (suite *OrganizationControllerTestSuite) TestGetOrganizationsEmptyResult() {
	expectedOrgs := []*models.Organization{}
	
	suite.mockService.On("GetOrganizations", "").Return(expectedOrgs, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/organization", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/organization", suite.orgController.GetOrganizations)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	
	// Check empty organizations array
	data, ok := response.Data.(map[string]interface{})
	assert.True(suite.T(), ok)
	
	organizations, ok := data["organizations"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), organizations, 0)
}

// TestFormatValidationErrors tests the validation error formatting function
func (suite *OrganizationControllerTestSuite) TestFormatValidationErrors() {
	testCases := []struct {
		name     string
		mockReq  interface{}
		expected []string
	}{
		{
			name: "Required field validation",
			mockReq: struct {
				Name string `validate:"required"`
			}{Name: ""},
			expected: []string{"Name is required"},
		},
		{
			name: "Min length validation",
			mockReq: struct {
				Name string `validate:"min=3"`
			}{Name: "ab"},
			expected: []string{"Name must be at least 3 characters/items"},
		},
		{
			name: "Max length validation",
			mockReq: struct {
				Name string `validate:"max=5"`
			}{Name: "toolong"},
			expected: []string{"Name must be at most 5 characters/items"},
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.orgController.validator.Struct(tc.mockReq)
			if err != nil {
				errorMsg := suite.orgController.formatValidationErrors(err)
				for _, expected := range tc.expected {
					assert.Contains(suite.T(), errorMsg, expected)
				}
			}
		})
	}
}