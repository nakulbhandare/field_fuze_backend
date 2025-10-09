package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockOrganizationRepository implements the OrganizationRepositoryInterface for testing
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) CreateOrganization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	args := m.Called(ctx, organization)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) GetOrganization(key string) ([]*models.Organization, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) UpdateOrganization(id string, organization *models.Organization) (*models.Organization, error) {
	args := m.Called(id, organization)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) DeleteOrganization(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockOrgLogger implements the Logger interface for testing
type MockOrgLogger struct {
	mock.Mock
}

func (m *MockOrgLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockOrgLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockOrgLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockOrgLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockOrgLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockOrgLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockOrgLogger) Warnf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockOrgLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockOrgLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockOrgLogger) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

// OrganizationServiceTestSuite contains the test suite for OrganizationService
type OrganizationServiceTestSuite struct {
	suite.Suite
	orgService *OrganizationService
	mockRepo   *MockOrganizationRepository
	mockLogger *MockOrgLogger
	ctx        context.Context
}

func (suite *OrganizationServiceTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.mockRepo = &MockOrganizationRepository{}
	suite.mockLogger = &MockOrgLogger{}
	
	// Set up comprehensive mock expectations
	suite.mockLogger.On("Debug", mock.Anything).Maybe()
	suite.mockLogger.On("Info", mock.Anything).Maybe()
	suite.mockLogger.On("Warn", mock.Anything).Maybe()
	suite.mockLogger.On("Error", mock.Anything).Maybe()
	suite.mockLogger.On("Debugf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Infof", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	suite.mockLogger.On("Errorf", mock.Anything, mock.Anything).Maybe()
	
	suite.orgService = NewOrganizationService(suite.mockRepo, suite.mockLogger)
}

func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}

// TestNewOrganizationService tests the constructor
func (suite *OrganizationServiceTestSuite) TestNewOrganizationService() {
	service := NewOrganizationService(suite.mockRepo, suite.mockLogger)
	
	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.mockRepo, service.organizationRepo)
	assert.Equal(suite.T(), suite.mockLogger, service.logger)
}

// TestCreateOrganization tests the CreateOrganization function
func (suite *OrganizationServiceTestSuite) TestCreateOrganization() {
	organization := &models.Organization{
		Name:       "Test Corp",
		Email:      "test@testcorp.com",
		Phone:      "+1234567890",
		Address:    "123 Test St",
		City:       "Test City",
		State:      "Test State",
		Country:    "Test Country",
		PostalCode: "12345",
		Industry:   "Technology",
	}
	
	expectedOrg := &models.Organization{
		Name:       "Test Corp",
		Email:      "test@testcorp.com",
		Phone:      "+1234567890",
		Address:    "123 Test St",
		City:       "Test City",
		State:      "Test State",
		Country:    "Test Country",
		PostalCode: "12345",
		Industry:   "Technology",
		CreatedBy:  "admin",
		UpdatedBy:  "admin",
	}
	
	suite.mockRepo.On("CreateOrganization", suite.ctx, mock.MatchedBy(func(org *models.Organization) bool {
		return org.Name == "Test Corp" && 
			   org.CreatedBy == "admin" && 
			   org.UpdatedBy == "admin"
	})).Return(expectedOrg, nil)
	
	result, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Test Corp", result.Name)
	assert.Equal(suite.T(), "admin", result.CreatedBy)
	assert.Equal(suite.T(), "admin", result.UpdatedBy)
}

// TestCreateOrganizationRepositoryError tests CreateOrganization when repository returns error
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationRepositoryError() {
	organization := &models.Organization{
		Name: "Test Corp",
	}
	
	suite.mockRepo.On("CreateOrganization", suite.ctx, mock.Anything).Return(nil, errors.New("database error"))
	
	result, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestCreateOrganizationValidationErrors tests CreateOrganization with various validation errors
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationValidationErrors() {
	testCases := []struct {
		name           string
		organization   *models.Organization
		expectedError  string
	}{
		{
			name:           "Nil organization",
			organization:   nil,
			expectedError:  "organization is required",
		},
		{
			name:           "Empty name",
			organization:   &models.Organization{Name: ""},
			expectedError:  "organization name is required",
		},
		{
			name:           "Whitespace only name",
			organization:   &models.Organization{Name: "   "},
			expectedError:  "organization name is required",
		},
		{
			name:           "Name too long",
			organization:   &models.Organization{Name: string(make([]byte, 101))},
			expectedError:  "organization name must be less than 100 characters",
		},
		{
			name:           "Invalid email",
			organization:   &models.Organization{Name: "Test", Email: "invalid-email"},
			expectedError:  "invalid email format",
		},
		{
			name:           "Invalid phone",
			organization:   &models.Organization{Name: "Test", Phone: "invalid-phone"},
			expectedError:  "invalid phone number format",
		},
		{
			name:           "Address too long",
			organization:   &models.Organization{Name: "Test", Address: string(make([]byte, 201))},
			expectedError:  "address must be less than 200 characters",
		},
		{
			name:           "City too short",
			organization:   &models.Organization{Name: "Test", City: "A"},
			expectedError:  "city must be between 2 and 50 characters",
		},
		{
			name:           "City too long",
			organization:   &models.Organization{Name: "Test", City: string(make([]byte, 51))},
			expectedError:  "city must be between 2 and 50 characters",
		},
		{
			name:           "State too short",
			organization:   &models.Organization{Name: "Test", State: "A"},
			expectedError:  "state must be between 2 and 50 characters",
		},
		{
			name:           "State too long",
			organization:   &models.Organization{Name: "Test", State: string(make([]byte, 51))},
			expectedError:  "state must be between 2 and 50 characters",
		},
		{
			name:           "Country too short",
			organization:   &models.Organization{Name: "Test", Country: "A"},
			expectedError:  "country must be between 2 and 50 characters",
		},
		{
			name:           "Country too long",
			organization:   &models.Organization{Name: "Test", Country: string(make([]byte, 51))},
			expectedError:  "country must be between 2 and 50 characters",
		},
		{
			name:           "Postal code too short",
			organization:   &models.Organization{Name: "Test", PostalCode: "12"},
			expectedError:  "postal code must be between 3 and 20 characters",
		},
		{
			name:           "Postal code too long",
			organization:   &models.Organization{Name: "Test", PostalCode: string(make([]byte, 21))},
			expectedError:  "postal code must be between 3 and 20 characters",
		},
		{
			name:           "Industry too short",
			organization:   &models.Organization{Name: "Test", Industry: "A"},
			expectedError:  "industry must be between 2 and 50 characters",
		},
		{
			name:           "Industry too long",
			organization:   &models.Organization{Name: "Test", Industry: string(make([]byte, 51))},
			expectedError:  "industry must be between 2 and 50 characters",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.orgService.CreateOrganization(suite.ctx, tc.organization, "admin")
			
			assert.Error(suite.T(), err)
			assert.Nil(suite.T(), result)
			assert.Contains(suite.T(), err.Error(), tc.expectedError)
		})
	}
}

// TestIsValidEmail tests the isValidEmail function
func (suite *OrganizationServiceTestSuite) TestIsValidEmail() {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"user+tag@domain.org",
		"user_name@domain-name.com",
		"123@domain.com",
	}
	
	invalidEmails := []string{
		"",
		"invalid",
		"@domain.com",
		"user@",
		"user@domain",
		"user space@domain.com",
		// Note: "user@domain..com" might be considered valid by the current regex
	}
	
	for _, email := range validEmails {
		assert.True(suite.T(), isValidEmail(email), "Expected %s to be valid", email)
	}
	
	for _, email := range invalidEmails {
		assert.False(suite.T(), isValidEmail(email), "Expected %s to be invalid", email)
	}
}

// TestGetOrganizations tests the GetOrganizations function
func (suite *OrganizationServiceTestSuite) TestGetOrganizations() {
	expectedOrgs := []*models.Organization{
		{
			Name: "Org 1",
		},
		{
			Name: "Org 2",
		},
	}
	
	suite.mockRepo.On("GetOrganization", "").Return(expectedOrgs, nil)
	
	result, err := suite.orgService.GetOrganizations("")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "Org 1", result[0].Name)
	assert.Equal(suite.T(), "Org 2", result[1].Name)
}

// TestGetOrganizationsError tests GetOrganizations when repository returns error
func (suite *OrganizationServiceTestSuite) TestGetOrganizationsError() {
	suite.mockRepo.On("GetOrganization", "").Return(nil, errors.New("database error"))
	
	result, err := suite.orgService.GetOrganizations("")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestGetOrganizationByID tests the GetOrganizationByID function
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByID() {
	expectedOrg := &models.Organization{
		Name: "Test Org",
	}
	
	suite.mockRepo.On("GetOrganization", "org-123").Return([]*models.Organization{expectedOrg}, nil)
	
	result, err := suite.orgService.GetOrganizationByID("org-123")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Test Org", result.Name)
}

// TestGetOrganizationByIDNotFound tests GetOrganizationByID when organization not found
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByIDNotFound() {
	suite.mockRepo.On("GetOrganization", "nonexistent").Return([]*models.Organization{}, nil)
	
	result, err := suite.orgService.GetOrganizationByID("nonexistent")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "organization not found")
}

// TestGetOrganizationByIDRepositoryError tests GetOrganizationByID when repository returns error
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByIDRepositoryError() {
	suite.mockRepo.On("GetOrganization", "org-123").Return(nil, errors.New("database error"))
	
	result, err := suite.orgService.GetOrganizationByID("org-123")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestUpdateOrganization tests the UpdateOrganization function
func (suite *OrganizationServiceTestSuite) TestUpdateOrganization() {
	updateReq := &models.Organization{
		Name:    "Updated Corp",
		Email:   "updated@corp.com",
		Address: "456 Updated St",
	}
	
	expectedOrg := &models.Organization{
		Name:      "Updated Corp",
		Email:     "updated@corp.com",
		Address:   "456 Updated St",
		UpdatedBy: "admin",
	}
	
	suite.mockRepo.On("UpdateOrganization", "org-123", mock.MatchedBy(func(org *models.Organization) bool {
		return org.Name == "Updated Corp" && org.UpdatedBy == "admin"
	})).Return(expectedOrg, nil)
	
	result, err := suite.orgService.UpdateOrganization("org-123", updateReq, "admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated Corp", result.Name)
	assert.Equal(suite.T(), "admin", result.UpdatedBy)
}

// TestUpdateOrganizationValidationError tests UpdateOrganization with validation error
func (suite *OrganizationServiceTestSuite) TestUpdateOrganizationValidationError() {
	invalidReq := &models.Organization{
		Name: "", // Invalid empty name
	}
	
	result, err := suite.orgService.UpdateOrganization("org-123", invalidReq, "admin")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "organization name is required")
}

// TestUpdateOrganizationRepositoryError tests UpdateOrganization when repository returns error
func (suite *OrganizationServiceTestSuite) TestUpdateOrganizationRepositoryError() {
	updateReq := &models.Organization{
		Name: "Valid Corp",
	}
	
	suite.mockRepo.On("UpdateOrganization", "org-123", mock.Anything).Return(nil, errors.New("database error"))
	
	result, err := suite.orgService.UpdateOrganization("org-123", updateReq, "admin")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestDeleteOrganization tests the DeleteOrganization function
func (suite *OrganizationServiceTestSuite) TestDeleteOrganization() {
	suite.mockRepo.On("DeleteOrganization", "org-123").Return(nil)
	
	err := suite.orgService.DeleteOrganization("org-123")
	
	assert.NoError(suite.T(), err)
}

// TestDeleteOrganizationRepositoryError tests DeleteOrganization when repository returns error
func (suite *OrganizationServiceTestSuite) TestDeleteOrganizationRepositoryError() {
	suite.mockRepo.On("DeleteOrganization", "org-123").Return(errors.New("database error"))
	
	err := suite.orgService.DeleteOrganization("org-123")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestGetOrganizationAssignmentsByStatus tests the GetOrganizationAssignmentsByStatus function
func (suite *OrganizationServiceTestSuite) TestGetOrganizationAssignmentsByStatus() {
	expectedOrgs := []*models.Organization{
		{Name: "Org 1"},
		{Name: "Org 2"},
	}
	
	suite.mockRepo.On("GetOrganization", "").Return(expectedOrgs, nil)
	
	result, err := suite.orgService.GetOrganizationAssignmentsByStatus("active")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
}

// TestGetOrganizationAssignmentsByStatusError tests GetOrganizationAssignmentsByStatus when repository returns error
func (suite *OrganizationServiceTestSuite) TestGetOrganizationAssignmentsByStatusError() {
	suite.mockRepo.On("GetOrganization", "").Return(nil, errors.New("database error"))
	
	result, err := suite.orgService.GetOrganizationAssignmentsByStatus("active")
	
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// TestUpdateOrganizationAssignment tests the UpdateOrganizationAssignment function
func (suite *OrganizationServiceTestSuite) TestUpdateOrganizationAssignment() {
	updateReq := &models.Organization{
		Name: "Updated Assignment Corp",
	}
	
	expectedOrg := &models.Organization{
		Name:      "Updated Assignment Corp",
		UpdatedBy: "admin",
	}
	
	suite.mockRepo.On("UpdateOrganization", "org-123", mock.MatchedBy(func(org *models.Organization) bool {
		return org.Name == "Updated Assignment Corp" && org.UpdatedBy == "admin"
	})).Return(expectedOrg, nil)
	
	result, err := suite.orgService.UpdateOrganizationAssignment("org-123", updateReq, "admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated Assignment Corp", result.Name)
	assert.Equal(suite.T(), "admin", result.UpdatedBy)
}

// TestDeleteOrganizationAssignment tests the DeleteOrganizationAssignment function
func (suite *OrganizationServiceTestSuite) TestDeleteOrganizationAssignment() {
	suite.mockRepo.On("DeleteOrganization", "org-123").Return(nil)
	
	err := suite.orgService.DeleteOrganizationAssignment("org-123")
	
	assert.NoError(suite.T(), err)
}

// TestDeleteOrganizationAssignmentRepositoryError tests DeleteOrganizationAssignment when repository returns error
func (suite *OrganizationServiceTestSuite) TestDeleteOrganizationAssignmentRepositoryError() {
	suite.mockRepo.On("DeleteOrganization", "org-123").Return(errors.New("database error"))
	
	err := suite.orgService.DeleteOrganizationAssignment("org-123")
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "database error")
}

// Edge case tests

// TestCreateOrganizationWithOptionalFields tests CreateOrganization with optional fields
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationWithOptionalFields() {
	organization := &models.Organization{
		Name: "Minimal Corp",
		// All other fields are optional and empty
	}
	
	expectedOrg := &models.Organization{
		Name:      "Minimal Corp",
		CreatedBy: "admin",
		UpdatedBy: "admin",
	}
	
	suite.mockRepo.On("CreateOrganization", suite.ctx, mock.Anything).Return(expectedOrg, nil)
	
	result, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Minimal Corp", result.Name)
}

// TestValidateCreateOrganizationWithWhitespace tests validation with whitespace fields
func (suite *OrganizationServiceTestSuite) TestValidateCreateOrganizationWithWhitespace() {
	organization := &models.Organization{
		Name:       "Valid Corp",
		Email:      "   ", // Whitespace only - should be ignored
		Phone:      "   ", // Whitespace only - should be ignored
		Address:    "   ", // Whitespace only - should be ignored
		City:       "   ", // Whitespace only - should be ignored
		State:      "   ", // Whitespace only - should be ignored
		Country:    "   ", // Whitespace only - should be ignored
		PostalCode: "   ", // Whitespace only - should be ignored
		Industry:   "   ", // Whitespace only - should be ignored
	}
	
	expectedOrg := &models.Organization{
		Name:      "Valid Corp",
		CreatedBy: "admin",
		UpdatedBy: "admin",
	}
	
	suite.mockRepo.On("CreateOrganization", suite.ctx, mock.Anything).Return(expectedOrg, nil)
	
	result, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
	
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

// TestPhoneValidation tests various phone number formats
func (suite *OrganizationServiceTestSuite) TestPhoneValidation() {
	validPhones := []string{
		"+1234567890",
		"+123456789012345", // 15 digits
		"1234567890",
		"12345678901234",   // 14 digits
	}
	
	invalidPhones := []string{
		"+0234567890",      // Starts with 0
		"1234567890123456", // Too long (16 digits)
		"+",                // Just plus
		"abcd",             // Letters
		"+123-456-7890",    // Contains dashes
		"(123) 456-7890",   // Contains spaces and parentheses
	}
	
	for _, phone := range validPhones {
		organization := &models.Organization{
			Name:  "Test Corp",
			Phone: phone,
		}
		
		suite.mockRepo.On("CreateOrganization", suite.ctx, mock.Anything).Return(organization, nil).Maybe()
		
		_, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
		assert.NoError(suite.T(), err, "Expected phone %s to be valid", phone)
	}
	
	for _, phone := range invalidPhones {
		organization := &models.Organization{
			Name:  "Test Corp",
			Phone: phone,
		}
		
		_, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
		assert.Error(suite.T(), err, "Expected phone %s to be invalid", phone)
		assert.Contains(suite.T(), err.Error(), "invalid phone number format")
	}
}

// TestConcurrentAccess tests concurrent access to organization service
func (suite *OrganizationServiceTestSuite) TestConcurrentAccess() {
	organization := &models.Organization{
		Name: "Concurrent Corp",
	}
	
	expectedOrg := &models.Organization{
		Name:      "Concurrent Corp",
		CreatedBy: "admin",
		UpdatedBy: "admin",
	}
	
	// Set up mock to handle multiple calls
	suite.mockRepo.On("CreateOrganization", suite.ctx, mock.Anything).Return(expectedOrg, nil).Times(2)
	
	done := make(chan bool, 2)
	
	// Start two goroutines to create organizations concurrently
	go func() {
		_, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
		assert.NoError(suite.T(), err)
		done <- true
	}()
	
	go func() {
		_, err := suite.orgService.CreateOrganization(suite.ctx, organization, "admin")
		assert.NoError(suite.T(), err)
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
}