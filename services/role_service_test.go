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

// MockRoleRepository implements the RoleRepositoryInterface for testing
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRoleAssignment(ctx context.Context, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error) {
	args := m.Called(ctx, roleAssignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleRepository) GetRoleAssignments(id string) ([]*models.RoleAssignment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleRepository) GetRole(name string) ([]*models.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

func (m *MockRoleRepository) UpdateRole(id string, role *models.Role) (*models.Role, error) {
	args := m.Called(id, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) DeleteRole(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error) {
	args := m.Called(status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleRepository) UpdateRoleAssignment(id string, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error) {
	args := m.Called(id, roleAssignment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleAssignment), args.Error(1)
}

func (m *MockRoleRepository) DeleteRoleAssignment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// RoleServiceTestSuite defines a test suite for RoleService functions
type RoleServiceTestSuite struct {
	suite.Suite
	ctx         context.Context
	mockRepo    *MockRoleRepository
	mockLogger  *MockLogger
	roleService *RoleService
}

// SetupTest runs before each test
func (suite *RoleServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.mockRepo = &MockRoleRepository{}
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
	
	suite.roleService = NewRoleService(suite.mockRepo, suite.mockLogger)
}

// TearDownTest runs after each test
func (suite *RoleServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// TestNewRoleService tests the NewRoleService function
func (suite *RoleServiceTestSuite) TestNewRoleService() {
	service := NewRoleService(suite.mockRepo, suite.mockLogger)
	
	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.mockRepo, service.roleRepo)
	assert.Equal(suite.T(), suite.mockLogger, service.logger)
}

// TestCreateRole tests the CreateRole function with valid input
func (suite *RoleServiceTestSuite) TestCreateRole() {
	roleAssignment := &models.RoleAssignment{
		RoleName:    "Admin",
		Level:       10,
		Permissions: []string{"read", "write", "delete", "admin"},
		Context: map[string]string{
			"department": "engineering",
		},
	}
	
	expectedRole := &models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "Admin",
		Level:       10,
		Permissions: []string{"read", "write", "delete", "admin"},
		Context: map[string]string{
			"department": "engineering",
		},
		AssignedAt: time.Now(),
	}
	
	suite.mockRepo.On("CreateRoleAssignment", suite.ctx, mock.MatchedBy(func(r *models.RoleAssignment) bool {
		return r.RoleName == "Admin" && 
			   r.Level == 10 && 
			   len(r.Permissions) == 4
	})).Return(expectedRole, nil)
	
	result, err := suite.roleService.CreateRole(suite.ctx, roleAssignment, "admin-user")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedRole.RoleID, result.RoleID)
	assert.Equal(suite.T(), "Admin", result.RoleName)
	assert.Equal(suite.T(), 10, result.Level)
}

// TestCreateRoleWithWhitespace tests CreateRole with whitespace trimming
func (suite *RoleServiceTestSuite) TestCreateRoleWithWhitespace() {
	roleAssignment := &models.RoleAssignment{
		RoleName:    "  Admin  ",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	expectedRole := &models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "Admin",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	suite.mockRepo.On("CreateRoleAssignment", suite.ctx, mock.MatchedBy(func(r *models.RoleAssignment) bool {
		return r.RoleName == "Admin" && r.Level == 5
	})).Return(expectedRole, nil)
	
	result, err := suite.roleService.CreateRole(suite.ctx, roleAssignment, "admin-user")
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Admin", result.RoleName)
}

// TestCreateRoleValidationErrors tests CreateRole with various validation errors
func (suite *RoleServiceTestSuite) TestCreateRoleValidationErrors() {
	testCases := []struct {
		name        string
		role        *models.RoleAssignment
		expectedErr string
	}{
		{
			name:        "Nil role assignment",
			role:        nil,
			expectedErr: "role assignment is required",
		},
		{
			name: "Empty role name",
			role: &models.RoleAssignment{
				RoleName:    "",
				Level:       5,
				Permissions: []string{"read"},
			},
			expectedErr: "role name is required",
		},
		{
			name: "Role name too long",
			role: &models.RoleAssignment{
				RoleName:    strings.Repeat("A", 101),
				Level:       5,
				Permissions: []string{"read"},
			},
			expectedErr: "role name must be less than 100 characters",
		},
		{
			name: "Level too low",
			role: &models.RoleAssignment{
				RoleName:    "TestRole",
				Level:       0,
				Permissions: []string{"read"},
			},
			expectedErr: "role level must be between 1 and 10",
		},
		{
			name: "Level too high",
			role: &models.RoleAssignment{
				RoleName:    "TestRole",
				Level:       11,
				Permissions: []string{"read"},
			},
			expectedErr: "role level must be between 1 and 10",
		},
		{
			name: "No permissions",
			role: &models.RoleAssignment{
				RoleName:    "TestRole",
				Level:       5,
				Permissions: []string{},
			},
			expectedErr: "at least one permission is required",
		},
		{
			name: "Empty permission",
			role: &models.RoleAssignment{
				RoleName:    "TestRole",
				Level:       5,
				Permissions: []string{"read", "", "write"},
			},
			expectedErr: "permission cannot be empty",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.roleService.CreateRole(suite.ctx, tc.role, "admin-user")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestCreateRoleRepositoryError tests CreateRole when repository returns error
func (suite *RoleServiceTestSuite) TestCreateRoleRepositoryError() {
	roleAssignment := &models.RoleAssignment{
		RoleName:    "Admin",
		Level:       10,
		Permissions: []string{"read", "write"},
	}
	
	suite.mockRepo.On("CreateRoleAssignment", suite.ctx, mock.Anything).Return(nil, errors.New("repository error"))
	
	result, err := suite.roleService.CreateRole(suite.ctx, roleAssignment, "admin-user")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetRoleAssignments tests the GetRoleAssignments function
func (suite *RoleServiceTestSuite) TestGetRoleAssignments() {
	expectedRoles := []*models.RoleAssignment{
		{
			RoleID:   "role-1",
			RoleName: "Admin",
			Level:    10,
			Permissions: []string{"admin"},
		},
		{
			RoleID:   "role-2",
			RoleName: "User",
			Level:    1,
			Permissions: []string{"read"},
		},
	}
	
	suite.mockRepo.On("GetRoleAssignments", "").Return(expectedRoles, nil)
	
	result, err := suite.roleService.GetRoleAssignments()
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), expectedRoles, result)
}

// TestGetRoleAssignmentsError tests GetRoleAssignments when repository returns error
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentsError() {
	suite.mockRepo.On("GetRoleAssignments", "").Return(nil, errors.New("repository error"))
	
	result, err := suite.roleService.GetRoleAssignments()
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetRoleAssignmentByID tests the GetRoleAssignmentByID function
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentByID() {
	expectedRole := &models.RoleAssignment{
		RoleID:   "role-123",
		RoleName: "Admin",
		Level:    10,
		Permissions: []string{"admin"},
	}
	
	suite.mockRepo.On("GetRoleAssignments", "role-123").Return([]*models.RoleAssignment{expectedRole}, nil)
	
	result, err := suite.roleService.GetRoleAssignmentByID("role-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedRole, result)
}

// TestGetRoleAssignmentByIDValidationErrors tests GetRoleAssignmentByID with validation errors
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentByIDValidationErrors() {
	testCases := []struct {
		name        string
		roleID      string
		expectedErr string
	}{
		{
			name:        "Empty role ID",
			roleID:      "",
			expectedErr: "role assignment ID is required",
		},
		{
			name:        "Whitespace only role ID",
			roleID:      "   ",
			expectedErr: "role assignment ID is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.roleService.GetRoleAssignmentByID(tc.roleID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestGetRoleAssignmentByIDNotFound tests GetRoleAssignmentByID when role is not found
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentByIDNotFound() {
	suite.mockRepo.On("GetRoleAssignments", "non-existent").Return([]*models.RoleAssignment{}, nil)
	
	result, err := suite.roleService.GetRoleAssignmentByID("non-existent")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "role assignment not found")
}

// TestGetRoleByName tests the GetRoleByName function
func (suite *RoleServiceTestSuite) TestGetRoleByName() {
	expectedRole := &models.Role{
		ID:          "role-123",
		Name:        "Admin",
		Description: "Administrator role",
		Level:       10,
	}
	
	suite.mockRepo.On("GetRole", "Admin").Return([]*models.Role{expectedRole}, nil)
	
	result, err := suite.roleService.GetRoleByName("Admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedRole, result)
}

// TestGetRoleByNameValidationErrors tests GetRoleByName with validation errors
func (suite *RoleServiceTestSuite) TestGetRoleByNameValidationErrors() {
	testCases := []struct {
		name        string
		roleName    string
		expectedErr string
	}{
		{
			name:        "Empty role name",
			roleName:    "",
			expectedErr: "role name is required",
		},
		{
			name:        "Whitespace only role name",
			roleName:    "   ",
			expectedErr: "role name is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.roleService.GetRoleByName(tc.roleName)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestGetRoleByNameNotFound tests GetRoleByName when role is not found
func (suite *RoleServiceTestSuite) TestGetRoleByNameNotFound() {
	suite.mockRepo.On("GetRole", "NonExistent").Return([]*models.Role{}, nil)
	
	result, err := suite.roleService.GetRoleByName("NonExistent")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "role not found")
}

// TestUpdateRole tests the UpdateRole function
func (suite *RoleServiceTestSuite) TestUpdateRole() {
	level := 8
	updateRequest := &models.UpdateRoleRequest{
		Name:        "Updated Admin",
		Description: "Updated administrator role",
		Level:       &level,
		Permissions: []string{"read", "write", "delete"},
		Status:      models.RoleStatusActive,
	}
	
	expectedRole := &models.Role{
		ID:          "role-123",
		Name:        "Updated Admin",
		Description: "Updated administrator role",
		Level:       8,
		Permissions: []string{"read", "write", "delete"},
		Status:      models.RoleStatusActive,
		UpdatedAt:   time.Now(),
	}
	
	suite.mockRepo.On("UpdateRole", "role-123", mock.MatchedBy(func(r *models.Role) bool {
		return r.Name == "Updated Admin" && 
			   r.Description == "Updated administrator role" && 
			   r.Level == 8
	})).Return(expectedRole, nil)
	
	result, err := suite.roleService.UpdateRole("role-123", updateRequest, "admin-user")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated Admin", result.Name)
	assert.Equal(suite.T(), "Updated administrator role", result.Description)
	assert.Equal(suite.T(), 8, result.Level)
}

// TestUpdateRoleWithWhitespace tests UpdateRole with whitespace trimming
func (suite *RoleServiceTestSuite) TestUpdateRoleWithWhitespace() {
	updateRequest := &models.UpdateRoleRequest{
		Name:        "  Updated Admin  ",
		Description: "  Updated description  ",
	}
	
	expectedRole := &models.Role{
		ID:          "role-123",
		Name:        "Updated Admin",
		Description: "Updated description",
	}
	
	suite.mockRepo.On("UpdateRole", "role-123", mock.MatchedBy(func(r *models.Role) bool {
		return r.Name == "Updated Admin" && r.Description == "Updated description"
	})).Return(expectedRole, nil)
	
	result, err := suite.roleService.UpdateRole("role-123", updateRequest, "admin-user")
	
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Admin", result.Name)
	assert.Equal(suite.T(), "Updated description", result.Description)
}

// TestUpdateRoleValidationErrors tests UpdateRole with validation errors
func (suite *RoleServiceTestSuite) TestUpdateRoleValidationErrors() {
	level11 := 11
	level0 := 0
	
	testCases := []struct {
		name        string
		roleID      string
		request     *models.UpdateRoleRequest
		expectedErr string
	}{
		{
			name:        "Empty role ID",
			roleID:      "",
			request:     &models.UpdateRoleRequest{Name: "Test"},
			expectedErr: "role ID is required",
		},
		{
			name:   "Nil request",
			roleID: "role-123",
			request: nil,
			expectedErr: "update role request is required",
		},
		{
			name:   "Name too long",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Name: strings.Repeat("A", 101),
			},
			expectedErr: "role name must be less than 100 characters",
		},
		{
			name:   "Description too long",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Description: strings.Repeat("B", 501),
			},
			expectedErr: "role description must be less than 500 characters",
		},
		{
			name:   "Level too high",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Level: &level11,
			},
			expectedErr: "role level must be between 1 and 10",
		},
		{
			name:   "Level too low",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Level: &level0,
			},
			expectedErr: "role level must be between 1 and 10",
		},
		{
			name:   "Invalid status",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Status: "invalid-status",
			},
			expectedErr: "invalid status: invalid-status",
		},
		{
			name:   "Empty permission",
			roleID: "role-123",
			request: &models.UpdateRoleRequest{
				Permissions: []string{"read", "", "write"},
			},
			expectedErr: "permission cannot be empty",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.roleService.UpdateRole(tc.roleID, tc.request, "admin-user")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestDeleteRole tests the DeleteRole function
func (suite *RoleServiceTestSuite) TestDeleteRole() {
	suite.mockRepo.On("DeleteRole", "role-123").Return(nil)
	
	err := suite.roleService.DeleteRole("role-123")
	
	assert.NoError(suite.T(), err)
}

// TestDeleteRoleValidationError tests DeleteRole with validation error
func (suite *RoleServiceTestSuite) TestDeleteRoleValidationError() {
	err := suite.roleService.DeleteRole("")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "role ID is required")
}

// TestDeleteRoleRepositoryError tests DeleteRole when repository returns error
func (suite *RoleServiceTestSuite) TestDeleteRoleRepositoryError() {
	suite.mockRepo.On("DeleteRole", "role-123").Return(errors.New("repository error"))
	
	err := suite.roleService.DeleteRole("role-123")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestGetRoleAssignmentsByStatus tests the GetRoleAssignmentsByStatus function
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentsByStatus() {
	expectedRoles := []*models.RoleAssignment{
		{
			RoleID:      "role-1",
			RoleName:    "Test Role 1",
			Level:       5,
			Permissions: []string{"read"},
		},
		{
			RoleID:      "role-2",
			RoleName:    "Test Role 2",
			Level:       6,
			Permissions: []string{"write"},
		},
	}
	
	suite.mockRepo.On("GetRoleAssignmentsByStatus", string(models.RoleStatusActive)).Return(expectedRoles, nil)
	
	result, err := suite.roleService.GetRoleAssignmentsByStatus(string(models.RoleStatusActive))
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), expectedRoles, result)
}

// TestGetRoleAssignmentsByStatusValidationError tests GetRoleAssignmentsByStatus with validation error
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentsByStatusValidationError() {
	_, err := suite.roleService.GetRoleAssignmentsByStatus("")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "status is required")
}

// TestGetRoleAssignmentsByStatusRepositoryError tests GetRoleAssignmentsByStatus when repository returns error
func (suite *RoleServiceTestSuite) TestGetRoleAssignmentsByStatusRepositoryError() {
	suite.mockRepo.On("GetRoleAssignmentsByStatus", "active").Return(nil, errors.New("repository error"))
	
	result, err := suite.roleService.GetRoleAssignmentsByStatus("active")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// TestUpdateRoleAssignment tests the UpdateRoleAssignment function
func (suite *RoleServiceTestSuite) TestUpdateRoleAssignment() {
	roleAssignment := &models.RoleAssignment{
		RoleName:    "Updated Admin",
		Level:       9,
		Permissions: []string{"read", "write", "delete"},
	}
	
	expectedRole := &models.RoleAssignment{
		RoleID:      "role-123",
		RoleName:    "Updated Admin",
		Level:       9,
		Permissions: []string{"read", "write", "delete"},
	}
	
	suite.mockRepo.On("UpdateRoleAssignment", "role-123", mock.MatchedBy(func(r *models.RoleAssignment) bool {
		return r.RoleID == "role-123" && 
			   r.RoleName == "Updated Admin" && 
			   r.Level == 9
	})).Return(expectedRole, nil)
	
	result, err := suite.roleService.UpdateRoleAssignment("role-123", roleAssignment, "admin-user")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "role-123", result.RoleID)
	assert.Equal(suite.T(), "Updated Admin", result.RoleName)
}

// TestUpdateRoleAssignmentValidationErrors tests UpdateRoleAssignment with validation errors
func (suite *RoleServiceTestSuite) TestUpdateRoleAssignmentValidationErrors() {
	testCases := []struct {
		name        string
		roleID      string
		assignment  *models.RoleAssignment
		expectedErr string
	}{
		{
			name:        "Empty role ID",
			roleID:      "",
			assignment:  &models.RoleAssignment{RoleName: "Test", Level: 5, Permissions: []string{"read"}},
			expectedErr: "role assignment ID is required",
		},
		{
			name:        "Nil assignment",
			roleID:      "role-123",
			assignment:  nil,
			expectedErr: "role assignment is required",
		},
		{
			name:   "Invalid assignment",
			roleID: "role-123",
			assignment: &models.RoleAssignment{
				RoleName:    "",
				Level:       5,
				Permissions: []string{"read"},
			},
			expectedErr: "role name is required",
		},
	}
	
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.roleService.UpdateRoleAssignment(tc.roleID, tc.assignment, "admin-user")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestDeleteRoleAssignment tests the DeleteRoleAssignment function
func (suite *RoleServiceTestSuite) TestDeleteRoleAssignment() {
	suite.mockRepo.On("DeleteRoleAssignment", "role-123").Return(nil)
	
	err := suite.roleService.DeleteRoleAssignment("role-123")
	
	assert.NoError(suite.T(), err)
}

// TestDeleteRoleAssignmentValidationError tests DeleteRoleAssignment with validation error
func (suite *RoleServiceTestSuite) TestDeleteRoleAssignmentValidationError() {
	err := suite.roleService.DeleteRoleAssignment("")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "role assignment ID is required")
}

// TestDeleteRoleAssignmentRepositoryError tests DeleteRoleAssignment when repository returns error
func (suite *RoleServiceTestSuite) TestDeleteRoleAssignmentRepositoryError() {
	suite.mockRepo.On("DeleteRoleAssignment", "role-123").Return(errors.New("repository error"))
	
	err := suite.roleService.DeleteRoleAssignment("role-123")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "repository error")
}

// Run the test suite
func TestRoleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RoleServiceTestSuite))
}

// Standalone tests for validation functions

func TestValidateCreateRoleAssignment(t *testing.T) {
	mockRepo := &MockRoleRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewRoleService(mockRepo, mockLogger)
	
	// Test valid role assignment
	validRole := &models.RoleAssignment{
		RoleName:    "Admin",
		Level:       5,
		Permissions: []string{"read", "write"},
	}
	
	err := service.validateCreateRoleAssignment(validRole)
	assert.NoError(t, err)
	
	// Test boundary cases for level
	for level := 1; level <= 10; level++ {
		role := &models.RoleAssignment{
			RoleName:    "TestRole",
			Level:       level,
			Permissions: []string{"read"},
		}
		err := service.validateCreateRoleAssignment(role)
		assert.NoError(t, err, "Level %d should be valid", level)
	}
	
	// Test boundary case for role name length
	maxLengthName := strings.Repeat("A", 100)
	role := &models.RoleAssignment{
		RoleName:    maxLengthName,
		Level:       5,
		Permissions: []string{"read"},
	}
	err = service.validateCreateRoleAssignment(role)
	assert.NoError(t, err)
	
	// Test multiple permissions
	multiPermissionRole := &models.RoleAssignment{
		RoleName:    "MultiRole",
		Level:       7,
		Permissions: []string{"read", "write", "delete", "admin", "manage"},
	}
	err = service.validateCreateRoleAssignment(multiPermissionRole)
	assert.NoError(t, err)
}

func TestValidateUpdateRoleRequest(t *testing.T) {
	mockRepo := &MockRoleRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewRoleService(mockRepo, mockLogger)
	
	// Test valid update request
	level := 5
	validRequest := &models.UpdateRoleRequest{
		Name:        "Updated Role",
		Description: "Updated description",
		Level:       &level,
		Permissions: []string{"read", "write"},
		Status:      models.RoleStatusActive,
	}
	
	err := service.validateUpdateRoleRequest(validRequest)
	assert.NoError(t, err)
	
	// Test all valid statuses
	validStatuses := []models.RoleStatus{
		models.RoleStatusActive,
		models.RoleStatusInactive,
		models.RoleStatusArchived,
	}
	
	for _, status := range validStatuses {
		request := &models.UpdateRoleRequest{Status: status}
		err := service.validateUpdateRoleRequest(request)
		assert.NoError(t, err, "Status should be valid: %s", status)
	}
	
	// Test empty fields are allowed in updates
	emptyFieldsRequest := &models.UpdateRoleRequest{
		Name:        "",
		Description: "",
	}
	
	err = service.validateUpdateRoleRequest(emptyFieldsRequest)
	assert.NoError(t, err)
	
	// Test boundary cases for lengths
	maxLengthName := strings.Repeat("A", 100)
	maxLengthDesc := strings.Repeat("B", 500)
	
	boundaryRequest := &models.UpdateRoleRequest{
		Name:        maxLengthName,
		Description: maxLengthDesc,
	}
	
	err = service.validateUpdateRoleRequest(boundaryRequest)
	assert.NoError(t, err)
	
	// Test level boundaries
	for level := 1; level <= 10; level++ {
		request := &models.UpdateRoleRequest{Level: &level}
		err := service.validateUpdateRoleRequest(request)
		assert.NoError(t, err, "Level %d should be valid", level)
	}
}

func TestRoleValidationEdgeCases(t *testing.T) {
	mockRepo := &MockRoleRepository{}
	mockLogger := &MockLogger{}
	mockLogger.On("Debug", mock.Anything).Return().Maybe()
	mockLogger.On("Debugf", mock.AnythingOfType("string"), mock.Anything).Return().Maybe()
	
	service := NewRoleService(mockRepo, mockLogger)
	
	// Test whitespace-only permission
	roleWithWhitespacePermission := &models.RoleAssignment{
		RoleName:    "TestRole",
		Level:       5,
		Permissions: []string{"read", "   ", "write"},
	}
	
	err := service.validateCreateRoleAssignment(roleWithWhitespacePermission)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission cannot be empty")
	
	// Test whitespace-only role name
	roleWithWhitespaceName := &models.RoleAssignment{
		RoleName:    "   ",
		Level:       5,
		Permissions: []string{"read"},
	}
	
	err = service.validateCreateRoleAssignment(roleWithWhitespaceName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")
	
	// Test nil permissions in update request
	requestWithNilPermissions := &models.UpdateRoleRequest{
		Permissions: []string{"read", "   ", "write"},
	}
	
	err = service.validateUpdateRoleRequest(requestWithNilPermissions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission cannot be empty")
}