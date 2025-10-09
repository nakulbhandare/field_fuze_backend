package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockLogger implements the logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

// MockUserRepository implements the UserRepositoryInterface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUser(key string) ([]*models.User, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(id string, user *models.User) (*models.User, error) {
	args := m.Called(id, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) AssignRoles(ctx context.Context, userID string, roleAssignments []models.RoleAssignment) (*models.User, error) {
	args := m.Called(ctx, userID, roleAssignments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) AddRoleToUser(ctx context.Context, userID string, roleAssignment models.RoleAssignment) (*models.User, error) {
	args := m.Called(ctx, userID, roleAssignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) (*models.User, error) {
	args := m.Called(ctx, userID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) (*models.User, error) {
	args := m.Called(ctx, userID, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// UserServiceTestSuite defines a test suite for UserService functions
type UserServiceTestSuite struct {
	suite.Suite
	ctx         context.Context
	mockRepo    *MockUserRepository
	mockLogger  *MockLogger
	userService *UserService
}

// SetupTest runs before each test
func (suite *UserServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.mockRepo = &MockUserRepository{}
	suite.mockLogger = &MockLogger{}
	
	// Mock logger calls that might be made
	suite.mockLogger.On("Debug", mock.Anything).Return().Maybe()
	suite.mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	suite.mockLogger.On("Info", mock.Anything).Return().Maybe()
	suite.mockLogger.On("Infof", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	suite.mockLogger.On("Error", mock.Anything).Return().Maybe()
	suite.mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	suite.mockLogger.On("Warn", mock.Anything).Return().Maybe()
	suite.mockLogger.On("Warnf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	suite.userService = NewUserService(suite.ctx, suite.mockRepo, suite.mockLogger)
}

// TearDownTest runs after each test
func (suite *UserServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// TestNewUserService tests the NewUserService function
func (suite *UserServiceTestSuite) TestNewUserService() {
	service := NewUserService(suite.ctx, suite.mockRepo, suite.mockLogger)
	
	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.ctx, service.ctx)
	assert.Equal(suite.T(), suite.mockRepo, service.repo)
	assert.Equal(suite.T(), suite.mockLogger, service.logger)
}

// TestCreateUser tests the CreateUser function with valid input
func (suite *UserServiceTestSuite) TestCreateUser() {
	user := &models.User{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
		Status:    models.UserStatusActive,
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "hashed-password",
		Status:    models.UserStatusActive,
		CreatedAt: time.Now(),
	}
	
	suite.mockRepo.On("CreateUser", suite.ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@example.com" && 
			   u.Username == "testuser" && 
			   u.FirstName == "Test" && 
			   u.LastName == "User"
	})).Return(expectedUser, nil)
	
	result, err := suite.userService.CreateUser(user)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), "test@example.com", result.Email)
	assert.Equal(suite.T(), "testuser", result.Username)
}

// TestCreateUserWithNormalization tests CreateUser with email/username normalization
func (suite *UserServiceTestSuite) TestCreateUserWithNormalization() {
	user := &models.User{
		Email:     "Test@Example.Com",  // Mixed case, no whitespace (passes validation)
		Username:  "TestUser",         // Mixed case, no whitespace (passes validation)
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",  // Should be normalized to lowercase
		Username:  "testuser",         // Should be normalized to lowercase
		FirstName: "Test",
		LastName:  "User",
		Status:    models.UserStatusActive,
	}
	
	suite.mockRepo.On("CreateUser", suite.ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "test@example.com" && 
			   u.Username == "testuser" && 
			   u.FirstName == "Test" && 
			   u.LastName == "User" &&
			   u.Password == "password123"
	})).Return(expectedUser, nil)
	
	result, err := suite.userService.CreateUser(user)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test@example.com", result.Email)
	assert.Equal(suite.T(), "testuser", result.Username)
	assert.Equal(suite.T(), "Test", result.FirstName)
	assert.Equal(suite.T(), "User", result.LastName)
}

// TestCreateUserValidationErrors tests CreateUser with various validation errors
func (suite *UserServiceTestSuite) TestCreateUserValidationErrors() {
	testCases := []struct {
		name        string
		user        *models.User
		expectedErr string
	}{
		{
			name:        "Nil user",
			user:        nil,
			expectedErr: "user is required",
		},
		{
			name: "Empty email",
			user: &models.User{
				Email:     "",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "email is required",
		},
		{
			name: "Empty username",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "username is required",
		},
		{
			name: "Empty first name",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "first name is required",
		},
		{
			name: "Empty last name",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "",
				Password:  "password123",
			},
			expectedErr: "last name is required",
		},
		{
			name: "Empty password",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "",
			},
			expectedErr: "password is required",
		},
		{
			name: "Invalid email format",
			user: &models.User{
				Email:     "invalid-email",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "invalid email format",
		},
		{
			name: "Invalid username format",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "ab", // Too short
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "username must be 3-30 characters and contain only letters, numbers, underscore, or hyphen",
		},
		{
			name: "Username with invalid characters",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "test@user", // Contains @
				FirstName: "Test",
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "username must be 3-30 characters and contain only letters, numbers, underscore, or hyphen",
		},
		{
			name: "Password too short",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  "User",
				Password:  "short",
			},
			expectedErr: "password must be at least 8 characters long",
		},
		{
			name: "First name too long",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: strings.Repeat("A", 51),
				LastName:  "User",
				Password:  "password123",
			},
			expectedErr: "first name must be less than 50 characters",
		},
		{
			name: "Last name too long",
			user: &models.User{
				Email:     "test@example.com",
				Username:  "testuser",
				FirstName: "Test",
				LastName:  strings.Repeat("B", 51),
				Password:  "password123",
			},
			expectedErr: "last name must be less than 50 characters",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.CreateUser(tc.user)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestCreateUserRepositoryError tests CreateUser when repository returns error
func (suite *UserServiceTestSuite) TestCreateUserRepositoryError() {
	user := &models.User{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
	}
	
	suite.mockRepo.On("CreateUser", suite.ctx, mock.Anything).Return(nil, errors.New("repository error"))
	
	result, err := suite.userService.CreateUser(user)
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetUsers tests the GetUsers function
func (suite *UserServiceTestSuite) TestGetUsers() {
	expectedUsers := []*models.User{
		{
			ID:       "user-1",
			Email:    "user1@example.com",
			Username: "user1",
		},
		{
			ID:       "user-2",
			Email:    "user2@example.com",
			Username: "user2",
		},
	}
	
	suite.mockRepo.On("GetUser", "").Return(expectedUsers, nil)
	
	result, err := suite.userService.GetUsers()
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), expectedUsers, result)
}

// TestGetUsersError tests GetUsers when repository returns error
func (suite *UserServiceTestSuite) TestGetUsersError() {
	suite.mockRepo.On("GetUser", "").Return(nil, errors.New("repository error"))
	
	result, err := suite.userService.GetUsers()
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetUserByID tests the GetUserByID function
func (suite *UserServiceTestSuite) TestGetUserByID() {
	expectedUser := &models.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Username: "testuser",
	}
	
	suite.mockRepo.On("GetUser", "user-123").Return([]*models.User{expectedUser}, nil)
	
	result, err := suite.userService.GetUserByID("user-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser, result)
}

// TestGetUserByIDValidationErrors tests GetUserByID with validation errors
func (suite *UserServiceTestSuite) TestGetUserByIDValidationErrors() {
	testCases := []struct {
		name        string
		userID      string
		expectedErr string
	}{
		{
			name:        "Empty user ID",
			userID:      "",
			expectedErr: "user ID is required",
		},
		{
			name:        "Whitespace only user ID",
			userID:      "   ",
			expectedErr: "user ID is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.GetUserByID(tc.userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestGetUserByIDNotFound tests GetUserByID when user is not found
func (suite *UserServiceTestSuite) TestGetUserByIDNotFound() {
	suite.mockRepo.On("GetUser", "non-existent").Return([]*models.User{}, nil)
	
	result, err := suite.userService.GetUserByID("non-existent")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

// TestGetUserByIDRepositoryError tests GetUserByID when repository returns error
func (suite *UserServiceTestSuite) TestGetUserByIDRepositoryError() {
	suite.mockRepo.On("GetUser", "user-123").Return(nil, errors.New("repository error"))
	
	result, err := suite.userService.GetUserByID("user-123")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetUserByEmail tests the GetUserByEmail function
func (suite *UserServiceTestSuite) TestGetUserByEmail() {
	expectedUser := &models.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Username: "testuser",
	}
	
	suite.mockRepo.On("GetUser", "test@example.com").Return([]*models.User{expectedUser}, nil)
	
	result, err := suite.userService.GetUserByEmail("test@example.com")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser, result)
}

// TestGetUserByEmailValidationErrors tests GetUserByEmail with validation errors
func (suite *UserServiceTestSuite) TestGetUserByEmailValidationErrors() {
	testCases := []struct {
		name        string
		email       string
		expectedErr string
	}{
		{
			name:        "Empty email",
			email:       "",
			expectedErr: "email is required",
		},
		{
			name:        "Whitespace only email",
			email:       "   ",
			expectedErr: "email is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.GetUserByEmail(tc.email)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestGetUserByEmailNotFound tests GetUserByEmail when user is not found
func (suite *UserServiceTestSuite) TestGetUserByEmailNotFound() {
	suite.mockRepo.On("GetUser", "notfound@example.com").Return([]*models.User{}, nil)
	
	result, err := suite.userService.GetUserByEmail("notfound@example.com")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "user not found")
}

// TestGetUserByUsername tests the GetUserByUsername function
func (suite *UserServiceTestSuite) TestGetUserByUsername() {
	expectedUser := &models.User{
		ID:       "user-123",
		Email:    "test@example.com",
		Username: "testuser",
	}
	
	suite.mockRepo.On("GetUser", "testuser").Return([]*models.User{expectedUser}, nil)
	
	result, err := suite.userService.GetUserByUsername("testuser")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser, result)
}

// TestGetUserByUsernameValidationErrors tests GetUserByUsername with validation errors
func (suite *UserServiceTestSuite) TestGetUserByUsernameValidationErrors() {
	testCases := []struct {
		name        string
		username    string
		expectedErr string
	}{
		{
			name:        "Empty username",
			username:    "",
			expectedErr: "username is required",
		},
		{
			name:        "Whitespace only username",
			username:    "   ",
			expectedErr: "username is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.GetUserByUsername(tc.username)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestUpdateUser tests the UpdateUser function
func (suite *UserServiceTestSuite) TestUpdateUser() {
	userUpdate := &models.User{
		FirstName: "Updated",
		LastName:  "Name",
		Status:    models.UserStatusInactive,
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Updated",
		LastName:  "Name",
		Status:    models.UserStatusInactive,
		UpdatedAt: time.Now(),
	}
	
	suite.mockRepo.On("UpdateUser", "user-123", mock.MatchedBy(func(u *models.User) bool {
		return u.FirstName == "Updated" && u.LastName == "Name"
	})).Return(expectedUser, nil)
	
	result, err := suite.userService.UpdateUser("user-123", userUpdate)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated", result.FirstName)
	assert.Equal(suite.T(), "Name", result.LastName)
}

// TestUpdateUserWithWhitespace tests UpdateUser with whitespace trimming
func (suite *UserServiceTestSuite) TestUpdateUserWithWhitespace() {
	userUpdate := &models.User{
		FirstName: "  Updated  ",
		LastName:  "  Name  ",
	}
	
	expectedUser := &models.User{
		ID:        "user-123",
		FirstName: "Updated",
		LastName:  "Name",
	}
	
	suite.mockRepo.On("UpdateUser", "user-123", mock.MatchedBy(func(u *models.User) bool {
		return u.FirstName == "Updated" && u.LastName == "Name"
	})).Return(expectedUser, nil)
	
	result, err := suite.userService.UpdateUser("user-123", userUpdate)
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", result.FirstName)
	assert.Equal(suite.T(), "Name", result.LastName)
}

// TestUpdateUserValidationErrors tests UpdateUser with validation errors
func (suite *UserServiceTestSuite) TestUpdateUserValidationErrors() {
	testCases := []struct {
		name        string
		userID      string
		user        *models.User
		expectedErr string
	}{
		{
			name:        "Empty user ID",
			userID:      "",
			user:        &models.User{FirstName: "Test"},
			expectedErr: "user ID is required",
		},
		{
			name:        "Nil user",
			userID:      "user-123",
			user:        nil,
			expectedErr: "user is required",
		},
		{
			name:   "First name too long",
			userID: "user-123",
			user: &models.User{
				FirstName: strings.Repeat("A", 51),
			},
			expectedErr: "first name must be less than 50 characters",
		},
		{
			name:   "Last name too long",
			userID: "user-123",
			user: &models.User{
				LastName: strings.Repeat("B", 51),
			},
			expectedErr: "last name must be less than 50 characters",
		},
		{
			name:   "Password too short",
			userID: "user-123",
			user: &models.User{
				Password: "short",
			},
			expectedErr: "password must be at least 8 characters long",
		},
		{
			name:   "Invalid status",
			userID: "user-123",
			user: &models.User{
				Status: "invalid-status",
			},
			expectedErr: "invalid status: invalid-status",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.UpdateUser(tc.userID, tc.user)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestAssignRolesToUser tests the AssignRolesToUser function
func (suite *UserServiceTestSuite) TestAssignRolesToUser() {
	roleAssignments := []models.RoleAssignment{
		{
			RoleID:   "role-1",
			RoleName: "Admin",
			Level:    10,
		},
		{
			RoleID:   "role-2",
			RoleName: "User",
			Level:    1,
		},
	}
	
	expectedUser := &models.User{
		ID:    "user-123",
		Roles: roleAssignments,
	}
	
	suite.mockRepo.On("AssignRoles", suite.ctx, "user-123", roleAssignments).Return(expectedUser, nil)
	
	result, err := suite.userService.AssignRolesToUser("user-123", roleAssignments)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Roles, 2)
}

// TestAssignRolesToUserValidationErrors tests AssignRolesToUser with validation errors
func (suite *UserServiceTestSuite) TestAssignRolesToUserValidationErrors() {
	testCases := []struct {
		name            string
		userID          string
		roleAssignments []models.RoleAssignment
		expectedErr     string
	}{
		{
			name:            "Empty user ID",
			userID:          "",
			roleAssignments: []models.RoleAssignment{{RoleID: "role-1"}},
			expectedErr:     "user ID is required",
		},
		{
			name:            "Empty role assignments",
			userID:          "user-123",
			roleAssignments: []models.RoleAssignment{},
			expectedErr:     "at least one role assignment is required",
		},
		{
			name:            "Nil role assignments",
			userID:          "user-123",
			roleAssignments: nil,
			expectedErr:     "at least one role assignment is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.AssignRolesToUser(tc.userID, tc.roleAssignments)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestAddRoleToUser tests the AddRoleToUser function
func (suite *UserServiceTestSuite) TestAddRoleToUser() {
	roleAssignment := models.RoleAssignment{
		RoleID:   "role-123",
		RoleName: "Admin",
		Level:    10,
	}
	
	expectedUser := &models.User{
		ID: "user-123",
		Roles: []models.RoleAssignment{roleAssignment},
	}
	
	suite.mockRepo.On("AddRoleToUser", suite.ctx, "user-123", roleAssignment).Return(expectedUser, nil)
	
	result, err := suite.userService.AddRoleToUser("user-123", roleAssignment)
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Roles, 1)
	assert.Equal(suite.T(), "role-123", result.Roles[0].RoleID)
}

// TestAddRoleToUserValidationErrors tests AddRoleToUser with validation errors
func (suite *UserServiceTestSuite) TestAddRoleToUserValidationErrors() {
	testCases := []struct {
		name           string
		userID         string
		roleAssignment models.RoleAssignment
		expectedErr    string
	}{
		{
			name:           "Empty user ID",
			userID:         "",
			roleAssignment: models.RoleAssignment{RoleID: "role-1", RoleName: "Admin"},
			expectedErr:    "user ID is required",
		},
		{
			name:           "Empty role ID",
			userID:         "user-123",
			roleAssignment: models.RoleAssignment{RoleID: "", RoleName: "Admin"},
			expectedErr:    "role ID is required",
		},
		{
			name:           "Empty role name",
			userID:         "user-123",
			roleAssignment: models.RoleAssignment{RoleID: "role-1", RoleName: ""},
			expectedErr:    "role name is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.AddRoleToUser(tc.userID, tc.roleAssignment)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestAssignRoleToUser tests the AssignRoleToUser function
func (suite *UserServiceTestSuite) TestAssignRoleToUser() {
	expectedUser := &models.User{
		ID: "user-123",
		Roles: []models.RoleAssignment{
			{RoleID: "role-123", RoleName: "Admin"},
		},
	}
	
	suite.mockRepo.On("AssignRoleToUser", suite.ctx, "user-123", "role-123").Return(expectedUser, nil)
	
	result, err := suite.userService.AssignRoleToUser("user-123", "role-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Roles, 1)
}

// TestAssignRoleToUserValidationErrors tests AssignRoleToUser with validation errors
func (suite *UserServiceTestSuite) TestAssignRoleToUserValidationErrors() {
	testCases := []struct {
		name        string
		userID      string
		roleID      string
		expectedErr string
	}{
		{
			name:        "Empty user ID",
			userID:      "",
			roleID:      "role-123",
			expectedErr: "user ID is required",
		},
		{
			name:        "Empty role ID",
			userID:      "user-123",
			roleID:      "",
			expectedErr: "role ID is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.AssignRoleToUser(tc.userID, tc.roleID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestRemoveRoleFromUser tests the RemoveRoleFromUser function
func (suite *UserServiceTestSuite) TestRemoveRoleFromUser() {
	expectedUser := &models.User{
		ID:    "user-123",
		Roles: []models.RoleAssignment{},
	}
	
	suite.mockRepo.On("RemoveRoleFromUser", suite.ctx, "user-123", "role-123").Return(expectedUser, nil)
	
	result, err := suite.userService.RemoveRoleFromUser("user-123", "role-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Roles, 0)
}

// TestRemoveRoleFromUserValidationErrors tests RemoveRoleFromUser with validation errors
func (suite *UserServiceTestSuite) TestRemoveRoleFromUserValidationErrors() {
	testCases := []struct {
		name        string
		userID      string
		roleID      string
		expectedErr string
	}{
		{
			name:        "Empty user ID",
			userID:      "",
			roleID:      "role-123",
			expectedErr: "user ID is required",
		},
		{
			name:        "Empty role ID",
			userID:      "user-123",
			roleID:      "",
			expectedErr: "role ID is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.userService.RemoveRoleFromUser(tc.userID, tc.roleID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestGetUsersByStatus tests the GetUsersByStatus function
func (suite *UserServiceTestSuite) TestGetUsersByStatus() {
	allUsers := []*models.User{
		{
			ID:     "user-1",
			Status: models.UserStatusActive,
		},
		{
			ID:     "user-2",
			Status: models.UserStatusInactive,
		},
		{
			ID:     "user-3",
			Status: models.UserStatusActive,
		},
	}
	
	suite.mockRepo.On("GetUser", "").Return(allUsers, nil)
	
	result, err := suite.userService.GetUsersByStatus(models.UserStatusActive)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "user-1", result[0].ID)
	assert.Equal(suite.T(), "user-3", result[1].ID)
}

// TestGetUsersByStatusValidationError tests GetUsersByStatus with validation error
func (suite *UserServiceTestSuite) TestGetUsersByStatusValidationError() {
	_, err := suite.userService.GetUsersByStatus("")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "status is required")
}

// TestGetUsersByStatusRepositoryError tests GetUsersByStatus when repository returns error
func (suite *UserServiceTestSuite) TestGetUsersByStatusRepositoryError() {
	suite.mockRepo.On("GetUser", "").Return(nil, errors.New("repository error"))
	
	result, err := suite.userService.GetUsersByStatus(models.UserStatusActive)
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetUsersByStatusNoMatches tests GetUsersByStatus when no users match status
func (suite *UserServiceTestSuite) TestGetUsersByStatusNoMatches() {
	allUsers := []*models.User{
		{
			ID:     "user-1",
			Status: models.UserStatusInactive,
		},
		{
			ID:     "user-2",
			Status: models.UserStatusInactive,
		},
	}
	
	suite.mockRepo.On("GetUser", "").Return(allUsers, nil)
	
	result, err := suite.userService.GetUsersByStatus(models.UserStatusActive)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
}

// Run the test suite
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// Standalone tests for validation functions

func TestValidateCreateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewUserService(ctx, mockRepo, mockLogger)
	
	// Test valid user
	validUser := &models.User{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		Password:  "password123",
	}
	
	err := service.validateCreateUser(validUser)
	assert.NoError(t, err)
	
	// Test edge cases for email validation
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user123@example123.com",
		"a@b.co",
	}
	
	for _, email := range validEmails {
		user := *validUser
		user.Email = email
		err := service.validateCreateUser(&user)
		assert.NoError(t, err, "Email should be valid: %s", email)
	}
	
	// Test edge cases for username validation
	validUsernames := []string{
		"abc",           // Minimum length
		"user123",       // With numbers
		"user_name",     // With underscore
		"user-name",     // With hyphen
		strings.Repeat("a", 30), // Maximum length
	}
	
	for _, username := range validUsernames {
		user := *validUser
		user.Username = username
		err := service.validateCreateUser(&user)
		assert.NoError(t, err, "Username should be valid: %s", username)
	}
}

func TestValidateUpdateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewUserService(ctx, mockRepo, mockLogger)
	
	// Test valid updates
	validUser := &models.User{
		FirstName: "Updated",
		LastName:  "Name",
		Password:  "newpassword123",
		Status:    models.UserStatusActive,
	}
	
	err := service.validateUpdateUser(validUser)
	assert.NoError(t, err)
	
	// Test all valid statuses
	validStatuses := []models.UserStatus{
		models.UserStatusActive,
		models.UserStatusInactive,
		models.UserStatusSuspended,
		models.UserStatusPendingVerification,
	}
	
	for _, status := range validStatuses {
		user := &models.User{Status: status}
		err := service.validateUpdateUser(user)
		assert.NoError(t, err, "Status should be valid: %s", status)
	}
	
	// Test empty fields are allowed in updates
	emptyFieldsUser := &models.User{
		FirstName: "",
		LastName:  "",
		Password:  "",
	}
	
	err = service.validateUpdateUser(emptyFieldsUser)
	assert.NoError(t, err)
}

func TestValidateRoleAssignment(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewUserService(ctx, mockRepo, mockLogger)
	
	// Test valid role assignment
	validRole := &models.RoleAssignment{
		RoleID:   "role-123",
		RoleName: "Admin",
		Level:    10,
	}
	
	err := service.validateRoleAssignment(validRole)
	assert.NoError(t, err)
	
	// Test nil role assignment
	err = service.validateRoleAssignment(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role assignment is required")
	
	// Test empty role ID
	invalidRole := &models.RoleAssignment{
		RoleID:   "",
		RoleName: "Admin",
	}
	
	err = service.validateRoleAssignment(invalidRole)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role ID is required")
	
	// Test empty role name
	invalidRole = &models.RoleAssignment{
		RoleID:   "role-123",
		RoleName: "",
	}
	
	err = service.validateRoleAssignment(invalidRole)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")
	
	// Test whitespace only fields
	invalidRole = &models.RoleAssignment{
		RoleID:   "   ",
		RoleName: "   ",
	}
	
	err = service.validateRoleAssignment(invalidRole)
	assert.Error(t, err)
}