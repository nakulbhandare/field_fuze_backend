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

// MockRoleService implements RoleServiceInterface for testing
type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) CreateRole(ctx context.Context, roleAssignment *models.RoleAssignment, createdBy string) (*models.RoleAssignment, error) {
	args := m.Called(ctx, roleAssignment, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleService) GetRoleAssignments() ([]*models.RoleAssignment, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleService) GetRoleAssignmentByID(id string) (*models.RoleAssignment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleService) GetRoleByName(name string) (*models.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) UpdateRole(id string, req *models.UpdateRoleRequest, updatedBy string) (*models.Role, error) {
	args := m.Called(id, req, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleService) DeleteRole(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleService) GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleService) UpdateRoleAssignment(id string, roleAssignment *models.RoleAssignment, updatedBy string) (*models.RoleAssignment, error) {
	args := m.Called(id, roleAssignment, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleService) DeleteRoleAssignment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// RoleControllerTestSuite contains the test suite for RoleController
type RoleControllerTestSuite struct {
	suite.Suite
	roleController *RoleController
	mockService    *MockRoleService
	mockLogger     *MockControllerLogger
	ctx            context.Context
	router         *gin.Engine
}

func (suite *RoleControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	suite.mockService = &MockRoleService{}
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
	
	suite.roleController = NewRoleController(suite.ctx, suite.mockService, suite.mockLogger)
	suite.router = gin.New()
}

func TestRoleControllerTestSuite(t *testing.T) {
	suite.Run(t, new(RoleControllerTestSuite))
}

// TestNewRoleController tests the constructor
func (suite *RoleControllerTestSuite) TestNewRoleController() {
	controller := NewRoleController(suite.ctx, suite.mockService, suite.mockLogger)
	
	assert.NotNil(suite.T(), controller)
	assert.Equal(suite.T(), suite.ctx, controller.ctx)
	assert.Equal(suite.T(), suite.mockService, controller.roleService)
	assert.Equal(suite.T(), suite.mockLogger, controller.logger)
}

// TestGetRoles tests the GetRoles endpoint
func (suite *RoleControllerTestSuite) TestGetRoles() {
	expectedRoles := []*models.RoleAssignment{
		{
			RoleID:      "role-1",
			RoleName:    "Admin",
			Level:       10,
			Permissions: []string{"admin", "read", "write"},
		},
		{
			RoleID:      "role-2",
			RoleName:    "User",
			Level:       1,
			Permissions: []string{"read"},
		},
	}
	
	suite.mockService.On("GetRoleAssignments").Return(expectedRoles, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/roles", suite.roleController.GetRoles)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.NotNil(suite.T(), response.Data)
}

// TestGetRolesWithFilters tests GetRoles with query filters
func (suite *RoleControllerTestSuite) TestGetRolesWithFilters() {
	expectedRoles := []*models.RoleAssignment{
		{
			RoleID:      "role-1",
			RoleName:    "Admin", 
			Level:       10,
			Permissions: []string{"admin"},
		},
	}
	
	suite.mockService.On("GetRoleAssignmentsByStatus", "active").Return(expectedRoles, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/roles?status=active", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/roles", suite.roleController.GetRoles)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestGetRolesServiceError tests GetRoles when service returns error
func (suite *RoleControllerTestSuite) TestGetRolesServiceError() {
	suite.mockService.On("GetRoleAssignments").Return(nil, errors.New("service error"))
	
	req, _ := http.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/roles", suite.roleController.GetRoles)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestCreateRole tests the CreateRole endpoint
func (suite *RoleControllerTestSuite) TestCreateRole() {
	roleReq := models.RoleAssignment{
		RoleName:    "Test Role",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	expectedRole := &models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "Test Role",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	suite.mockService.On("CreateRole", suite.ctx, mock.MatchedBy(func(role *models.RoleAssignment) bool {
		return role.RoleName == "Test Role" && role.Level == 5
	}), "admin").Return(expectedRole, nil)
	
	body, _ := json.Marshal(roleReq)
	req, _ := http.NewRequest(http.MethodPost, "/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	// Set user context for createdBy field
	ctx := req.Context()
	ctx = context.WithValue(ctx, "user", &models.JWTClaims{UserID: "admin"})
	req = req.WithContext(ctx)
	
	w := httptest.NewRecorder()
	suite.router.POST("/role", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.CreateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Role created successfully", response.Message)
}

// TestCreateRoleInvalidRequest tests CreateRole with invalid request
func (suite *RoleControllerTestSuite) TestCreateRoleInvalidRequest() {
	invalidReq := map[string]interface{}{
		"role_name": "", // Empty role name
		"level":     0,  // Invalid level
	}
	
	body, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest(http.MethodPost, "/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.POST("/role", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.CreateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestCreateRoleServiceError tests CreateRole when service returns error
func (suite *RoleControllerTestSuite) TestCreateRoleServiceError() {
	roleReq := models.RoleAssignment{
		RoleName:    "Test Role",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	suite.mockService.On("CreateRole", suite.ctx, mock.Anything, "admin").Return(nil, errors.New("service error"))
	
	body, _ := json.Marshal(roleReq)
	req, _ := http.NewRequest(http.MethodPost, "/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.POST("/role", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.CreateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestGetRole tests the GetRole endpoint
func (suite *RoleControllerTestSuite) TestGetRole() {
	expectedRole := &models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "Admin",
		Level:       10,
		Permissions: []string{"admin", "read", "write"},
	}
	
	suite.mockService.On("GetRoleAssignmentByID", "role-123").Return(expectedRole, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/role/role-123", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/role/:id", suite.roleController.GetRole)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.NotNil(suite.T(), response.Data)
}

// TestGetRoleNotFound tests GetRole when role not found
func (suite *RoleControllerTestSuite) TestGetRoleNotFound() {
	suite.mockService.On("GetRoleAssignmentByID", "nonexistent").Return(nil, errors.New("role not found"))
	
	req, _ := http.NewRequest(http.MethodGet, "/role/nonexistent", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/role/:id", suite.roleController.GetRole)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestUpdateRole tests the UpdateRole endpoint
func (suite *RoleControllerTestSuite) TestUpdateRole() {
	updateReq := models.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
		Level:       &[]int{7}[0], // Pointer to int
		Permissions: []string{"read", "write", "admin"},
		Status:      models.RoleStatusActive,
	}
	
	expectedRole := &models.Role{
		ID:          "role-123",
		Name:        "Updated Role",
		Description: "Updated description",
		Level:       7,
		Permissions: []string{"read", "write", "admin"},
		Status:      models.RoleStatusActive,
	}
	
	suite.mockService.On("UpdateRole", "role-123", mock.MatchedBy(func(req *models.UpdateRoleRequest) bool {
		return req.Name == "Updated Role" && *req.Level == 7
	}), "admin").Return(expectedRole, nil)
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/role/role-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.PUT("/role/:id", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.UpdateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestUpdateRoleInvalidRequest tests UpdateRole with invalid request
func (suite *RoleControllerTestSuite) TestUpdateRoleInvalidRequest() {
	req, _ := http.NewRequest(http.MethodPut, "/role/role-123", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.PUT("/role/:id", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.UpdateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestUpdateRoleServiceError tests UpdateRole when service returns error
func (suite *RoleControllerTestSuite) TestUpdateRoleServiceError() {
	updateReq := models.UpdateRoleRequest{
		Name: "Updated Role",
	}
	
	suite.mockService.On("UpdateRole", "role-123", mock.Anything, "admin").Return(nil, errors.New("service error"))
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/role/role-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.PUT("/role/:id", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.UpdateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestDeleteRole tests the DeleteRole endpoint
func (suite *RoleControllerTestSuite) TestDeleteRole() {
	suite.mockService.On("DeleteRole", "role-123").Return(nil)
	
	req, _ := http.NewRequest(http.MethodDelete, "/role/role-123", nil)
	w := httptest.NewRecorder()
	
	suite.router.DELETE("/role/:id", suite.roleController.DeleteRole)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Role deleted successfully", response.Message)
}

// TestDeleteRoleServiceError tests DeleteRole when service returns error
func (suite *RoleControllerTestSuite) TestDeleteRoleServiceError() {
	suite.mockService.On("DeleteRole", "role-123").Return(errors.New("service error"))
	
	req, _ := http.NewRequest(http.MethodDelete, "/role/role-123", nil)
	w := httptest.NewRecorder()
	
	suite.router.DELETE("/role/:id", suite.roleController.DeleteRole)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// Edge case tests

// TestGetRolesWithPagination tests GetRoles with pagination parameters
func (suite *RoleControllerTestSuite) TestGetRolesWithPagination() {
	expectedRoles := []*models.RoleAssignment{
		{RoleID: "role-1", RoleName: "Admin", Level: 10, Permissions: []string{"admin"}},
	}
	
	suite.mockService.On("GetRoleAssignments").Return(expectedRoles, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/roles?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/roles", suite.roleController.GetRoles)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestCreateRoleMalformedJSON tests CreateRole with malformed JSON
func (suite *RoleControllerTestSuite) TestCreateRoleMalformedJSON() {
	req, _ := http.NewRequest(http.MethodPost, "/role", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.POST("/role", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.CreateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestCreateRoleWithoutUser tests CreateRole without user context
func (suite *RoleControllerTestSuite) TestCreateRoleWithoutUser() {
	roleReq := models.RoleAssignment{
		RoleName:    "Test Role",
		Level:       5,
		Permissions: []string{"read"},
	}
	
	body, _ := json.Marshal(roleReq)
	req, _ := http.NewRequest(http.MethodPost, "/role", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.POST("/role", suite.roleController.CreateRole)
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestGetRoleEmptyID tests GetRole with empty ID
func (suite *RoleControllerTestSuite) TestGetRoleEmptyID() {
	req, _ := http.NewRequest(http.MethodGet, "/role/", nil)
	w := httptest.NewRecorder()
	
	suite.router.GET("/role/:id", suite.roleController.GetRole)
	suite.router.ServeHTTP(w, req)
	
	// Should not match the route, resulting in 404
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestUpdateRoleEmptyID tests UpdateRole with empty ID  
func (suite *RoleControllerTestSuite) TestUpdateRoleEmptyID() {
	updateReq := models.UpdateRoleRequest{Name: "Test"}
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPut, "/role/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	suite.router.PUT("/role/:id", func(c *gin.Context) {
		c.Set("jwt_claims", &models.JWTClaims{UserID: "admin"})
		suite.roleController.UpdateRole(c)
	})
	suite.router.ServeHTTP(w, req)
	
	// Should not match the route, resulting in 404
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestDeleteRoleEmptyID tests DeleteRole with empty ID
func (suite *RoleControllerTestSuite) TestDeleteRoleEmptyID() {
	req, _ := http.NewRequest(http.MethodDelete, "/role/", nil)
	w := httptest.NewRecorder()
	
	suite.router.DELETE("/role/:id", suite.roleController.DeleteRole)
	suite.router.ServeHTTP(w, req)
	
	// Should not match the route, resulting in 404
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}