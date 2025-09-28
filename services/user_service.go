package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"regexp"
	"strings"
)

type UserService struct {
	ctx    context.Context
	repo   repository.UserRepositoryInterface
	logger logger.Logger
}

func NewUserService(ctx context.Context, repo repository.UserRepositoryInterface, logger logger.Logger) *UserService {
	return &UserService{
		ctx:    ctx,
		repo:   repo,
		logger: logger,
	}
}

func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
	if err := s.validateCreateUser(user); err != nil {
		return nil, err
	}

	// Set system-generated fields
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Username = strings.ToLower(strings.TrimSpace(user.Username))
	user.FirstName = strings.TrimSpace(user.FirstName)
	user.LastName = strings.TrimSpace(user.LastName)

	return s.repo.CreateUser(s.ctx, user)
}

func (s *UserService) GetUsers() ([]*models.User, error) {
	return s.repo.GetUser("")
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("user ID is required")
	}

	users, err := s.repo.GetUser(id)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return users[0], nil
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}

	users, err := s.repo.GetUser(email)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return users[0], nil
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("username is required")
	}

	users, err := s.repo.GetUser(username)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return users[0], nil
}

func (s *UserService) UpdateUser(id string, user *models.User) (*models.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("user ID is required")
	}

	if err := s.validateUpdateUser(user); err != nil {
		return nil, err
	}

	// Sanitize input fields
	if user.FirstName != "" {
		user.FirstName = strings.TrimSpace(user.FirstName)
	}
	if user.LastName != "" {
		user.LastName = strings.TrimSpace(user.LastName)
	}

	return s.repo.UpdateUser(id, user)
}

func (s *UserService) AssignRolesToUser(userID string, roleAssignments []models.RoleAssignment) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	if len(roleAssignments) == 0 {
		return nil, errors.New("at least one role assignment is required")
	}

	return s.repo.AssignRoles(s.ctx, userID, roleAssignments)
}

func (s *UserService) AddRoleToUser(userID string, roleAssignment models.RoleAssignment) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	if err := s.validateRoleAssignment(&roleAssignment); err != nil {
		return nil, err
	}

	return s.repo.AddRoleToUser(s.ctx, userID, roleAssignment)
}

func (s *UserService) AssignRoleToUser(userID, roleID string) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	if strings.TrimSpace(roleID) == "" {
		return nil, errors.New("role ID is required")
	}

	return s.repo.AssignRoleToUser(s.ctx, userID, roleID)
}

func (s *UserService) RemoveRoleFromUser(userID, roleID string) (*models.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user ID is required")
	}

	if strings.TrimSpace(roleID) == "" {
		return nil, errors.New("role ID is required")
	}

	return s.repo.RemoveRoleFromUser(s.ctx, userID, roleID)
}

func (s *UserService) GetUsersByStatus(status models.UserStatus) ([]*models.User, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}

	// Get all users and filter by status
	users, err := s.repo.GetUser("")
	if err != nil {
		return nil, err
	}

	var filteredUsers []*models.User
	for _, user := range users {
		if user.Status == status {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return filteredUsers, nil
}

func (s *UserService) validateCreateUser(user *models.User) error {
	if user == nil {
		return errors.New("user is required")
	}

	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}

	if strings.TrimSpace(user.Username) == "" {
		return errors.New("username is required")
	}

	if strings.TrimSpace(user.FirstName) == "" {
		return errors.New("first name is required")
	}

	if strings.TrimSpace(user.LastName) == "" {
		return errors.New("last name is required")
	}

	if strings.TrimSpace(user.Password) == "" {
		return errors.New("password is required")
	}

	// Validate email format
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailPattern, user.Email); !matched {
		return errors.New("invalid email format")
	}

	// Validate username (alphanumeric, underscore, hyphen only)
	usernamePattern := `^[a-zA-Z0-9_-]{3,30}$`
	if matched, _ := regexp.MatchString(usernamePattern, user.Username); !matched {
		return errors.New("username must be 3-30 characters and contain only letters, numbers, underscore, or hyphen")
	}

	// Validate password strength
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Validate name lengths
	if len(user.FirstName) > 50 {
		return errors.New("first name must be less than 50 characters")
	}

	if len(user.LastName) > 50 {
		return errors.New("last name must be less than 50 characters")
	}

	return nil
}

func (s *UserService) validateUpdateUser(user *models.User) error {
	if user == nil {
		return errors.New("user is required")
	}

	if user.FirstName != "" && len(user.FirstName) > 50 {
		return errors.New("first name must be less than 50 characters")
	}

	if user.LastName != "" && len(user.LastName) > 50 {
		return errors.New("last name must be less than 50 characters")
	}

	if user.Password != "" && len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if user.Status != "" {
		validStatuses := []models.UserStatus{
			models.UserStatusActive,
			models.UserStatusInactive,
			models.UserStatusSuspended,
			models.UserStatusPendingVerification,
		}
		isValid := false
		for _, status := range validStatuses {
			if user.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid status: %s", user.Status)
		}
	}

	return nil
}

func (s *UserService) validateRoleAssignment(roleAssignment *models.RoleAssignment) error {
	if roleAssignment == nil {
		return errors.New("role assignment is required")
	}

	if strings.TrimSpace(roleAssignment.RoleID) == "" {
		return errors.New("role ID is required")
	}

	if strings.TrimSpace(roleAssignment.RoleName) == "" {
		return errors.New("role name is required")
	}

	return nil
}
