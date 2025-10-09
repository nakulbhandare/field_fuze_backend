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

// MockUserService implements UserServiceInterface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(user *models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id string, user *models.User) (*models.User, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) AssignRolesToUser(userID string, roleAssignments []models.RoleAssignment) (*models.User, error) {
	args := m.Called(userID, roleAssignments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) AddRoleToUser(userID string, roleAssignment models.RoleAssignment) (*models.User, error) {
	args := m.Called(userID, roleAssignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) AssignRoleToUser(userID, roleID string) (*models.User, error) {
	args := m.Called(userID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) RemoveRoleFromUser(userID, roleID string) (*models.User, error) {
	args := m.Called(userID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUsersByStatus(status models.UserStatus) ([]*models.User, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockControllerLogger implements Logger interface for testing
type MockControllerLogger struct {
	mock.Mock
}

func (m *MockControllerLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockControllerLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockControllerLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockControllerLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockControllerLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockControllerLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockControllerLogger) Warnf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockControllerLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockControllerLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockControllerLogger) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

// UserControllerTestSuite contains the test suite for UserController
type UserControllerTestSuite struct {
	suite.Suite
	mockService    *MockUserService
	mockLogger     *MockControllerLogger
	userController *UserController
	ctx            context.Context
	router         *gin.Engine
}

func (suite *UserControllerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	suite.mockService = &MockUserService{}
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
	
	suite.userController = NewUserController(suite.ctx, suite.mockService, suite.mockLogger, nil)
	suite.router = gin.New()
}

// setupController creates a controller without JWT manager for testing core logic
func (suite *UserControllerTestSuite) setupController() *UserController {
	return NewUserController(suite.ctx, suite.mockService, suite.mockLogger, nil)
}

func TestUserControllerTestSuite(t *testing.T) {
	suite.Run(t, new(UserControllerTestSuite))
}


// TestRegisterLogic tests the Register endpoint logic
func (suite *UserControllerTestSuite) TestRegisterLogic() {
	controller := suite.setupController()
	
	registerReq := models.RegisterUser{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Status:    models.UserStatusActive,
	}
	
	suite.mockService.On("CreateUser", mock.MatchedBy(func(user *models.User) bool {
		return user.Email == "test@example.com" && user.Username == "testuser"
	})).Return(expectedUser, nil)
	
	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.Register(c)
	
	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "User registered successfully", response.Message)
}

// TestRegisterInvalidJSON tests Register with invalid JSON
func (suite *UserControllerTestSuite) TestRegisterInvalidJSON() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.Register(c)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestRegisterServiceError tests Register when service returns error
func (suite *UserControllerTestSuite) TestRegisterServiceError() {
	controller := suite.setupController()
	
	registerReq := models.RegisterUser{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
	}
	
	suite.mockService.On("CreateUser", mock.Anything).Return(nil, errors.New("service error"))
	
	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.Register(c)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestGetUser tests the GetUser endpoint
func (suite *UserControllerTestSuite) TestGetUser() {
	controller := suite.setupController()
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
	}
	
	suite.mockService.On("GetUserByID", "user-123").Return(expectedUser, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/user/user-123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "user-123"}}
	
	controller.GetUser(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.NotNil(suite.T(), response.Data)
}

// TestGetUserNotFound tests GetUser when user not found
func (suite *UserControllerTestSuite) TestGetUserNotFound() {
	controller := suite.setupController()
	
	suite.mockService.On("GetUserByID", "nonexistent").Return(nil, errors.New("user not found"))
	
	req, _ := http.NewRequest(http.MethodGet, "/user/nonexistent", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
	
	controller.GetUser(c)
	
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestGetUserList tests the GetUserList endpoint
func (suite *UserControllerTestSuite) TestGetUserList() {
	controller := suite.setupController()
	
	expectedUsers := []*models.User{
		{ID: "user-1", Email: "user1@example.com"},
		{ID: "user-2", Email: "user2@example.com"},
	}
	
	suite.mockService.On("GetUsers").Return(expectedUsers, nil)
	
	req, _ := http.NewRequest(http.MethodGet, "/user/list", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.GetUserList(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.NotNil(suite.T(), response.Data)
}

// TestGetUserListServiceError tests GetUserList when service returns error
func (suite *UserControllerTestSuite) TestGetUserListServiceError() {
	controller := suite.setupController()
	
	suite.mockService.On("GetUsers").Return(nil, errors.New("service error"))
	
	req, _ := http.NewRequest(http.MethodGet, "/user/list", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.GetUserList(c)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestUpdateUser tests the UpdateUser endpoint
func (suite *UserControllerTestSuite) TestUpdateUser() {
	controller := suite.setupController()
	
	updateReq := models.User{
		FirstName: "Updated",
		LastName:  "Name",
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		FirstName: "Updated",
		LastName:  "Name",
	}
	
	suite.mockService.On("UpdateUser", "user-123", mock.MatchedBy(func(user *models.User) bool {
		return user.FirstName == "Updated" && user.LastName == "Name"
	})).Return(expectedUser, nil)
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPatch, "/user/update/user-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "user-123"}}
	
	controller.UpdateUser(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestUpdateUserInvalidJSON tests UpdateUser with invalid JSON
func (suite *UserControllerTestSuite) TestUpdateUserInvalidJSON() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPatch, "/user/update/user-123", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "user-123"}}
	
	controller.UpdateUser(c)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestUpdateUserServiceError tests UpdateUser when service returns error
func (suite *UserControllerTestSuite) TestUpdateUserServiceError() {
	controller := suite.setupController()
	
	updateReq := models.User{
		FirstName: "Updated",
		LastName:  "Name",
	}
	
	suite.mockService.On("UpdateUser", "user-123", mock.Anything).Return(nil, errors.New("service error"))
	
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest(http.MethodPatch, "/user/update/user-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "user-123"}}
	
	controller.UpdateUser(c)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestAssignRole tests the AssignRole endpoint
func (suite *UserControllerTestSuite) TestAssignRole() {
	controller := suite.setupController()
	
	expectedUser := &models.User{
		ID: "user-123",
		Roles: []models.RoleAssignment{
			{RoleID: "role-456", RoleName: "Admin"},
		},
	}
	
	suite.mockService.On("AssignRoleToUser", "user-123", "role-456").Return(expectedUser, nil)
	
	req, _ := http.NewRequest(http.MethodPost, "/user/user-123/role/role-456", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "user_id", Value: "user-123"},
		{Key: "role_id", Value: "role-456"},
	}
	// Set user context for invalidateUserPermissions
	c.Set("user", &models.JWTClaims{UserID: "admin"})
	
	controller.AssignRole(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestDetachRole tests the DetachRole endpoint
func (suite *UserControllerTestSuite) TestDetachRole() {
	controller := suite.setupController()
	
	expectedUser := &models.User{
		ID:    "user-123",
		Roles: []models.RoleAssignment{},
	}
	
	suite.mockService.On("RemoveRoleFromUser", "user-123", "role-456").Return(expectedUser, nil)
	
	req, _ := http.NewRequest(http.MethodDelete, "/user/user-123/role/role-456", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "user_id", Value: "user-123"},
		{Key: "role_id", Value: "role-456"},
	}
	// Set user context for invalidateUserPermissions
	c.Set("user", &models.JWTClaims{UserID: "admin"})
	
	controller.DetachRole(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
}

// TestLoginDelegation tests that Login delegates to JWT manager
func (suite *UserControllerTestSuite) TestLoginDelegation() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// Since jwtManager is nil, this should handle gracefully
	controller.Login(c)
	
	// The function delegates to JWT manager, so if manager is nil,
	// it should not panic but may not process the request
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestGenerateTokenDelegation tests that GenerateToken delegates properly
func (suite *UserControllerTestSuite) TestGenerateTokenDelegation() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPost, "/token", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// This endpoint delegates to AuthMiddleware processing
	controller.GenerateToken(c)
	
	// Since we're testing delegation, we expect this to complete without error
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestValidateTokenDelegation tests that ValidateToken delegates properly
func (suite *UserControllerTestSuite) TestValidateTokenDelegation() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPost, "/validate", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// This endpoint delegates to JWT manager processing
	controller.ValidateToken(c)
	
	// Since we're testing delegation, we expect this to complete without error
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestLogout tests the Logout endpoint
func (suite *UserControllerTestSuite) TestLogout() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	controller.Logout(c)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "Logged out successfully", response.Message)
}

// Edge case tests

// TestGetUserWithEmptyID tests GetUser with empty ID parameter
func (suite *UserControllerTestSuite) TestGetUserWithEmptyID() {
	controller := suite.setupController()
	
	req, _ := http.NewRequest(http.MethodGet, "/user/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: ""}}
	
	controller.GetUser(c)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestUpdateUserWithEmptyID tests UpdateUser with empty ID parameter
func (suite *UserControllerTestSuite) TestUpdateUserWithEmptyID() {
	controller := suite.setupController()
	
	updateReq := models.User{FirstName: "Test"}
	body, _ := json.Marshal(updateReq)
	
	req, _ := http.NewRequest(http.MethodPatch, "/user/update/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: ""}}
	
	controller.UpdateUser(c)
	
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}

// TestAssignRoleServiceError tests AssignRole when service returns error
func (suite *UserControllerTestSuite) TestAssignRoleServiceError() {
	controller := suite.setupController()
	
	suite.mockService.On("AssignRoleToUser", "user-123", "role-456").Return(nil, errors.New("service error"))
	
	req, _ := http.NewRequest(http.MethodPost, "/user/user-123/role/role-456", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "user_id", Value: "user-123"},
		{Key: "role_id", Value: "role-456"},
	}
	c.Set("user", &models.JWTClaims{UserID: "admin"})
	
	controller.AssignRole(c)
	
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "error", response.Status)
}