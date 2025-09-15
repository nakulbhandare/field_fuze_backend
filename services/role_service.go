package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"strings"
)

type RoleService struct {
	roleRepo *repository.RoleRepository
	logger   logger.Logger
}

func NewRoleService(roleRepo *repository.RoleRepository, logger logger.Logger) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
		logger:   logger,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *models.CreateRoleRequest, createdBy string) (*models.Role, error) {
	if err := s.validateCreateRoleRequest(req); err != nil {
		return nil, err
	}

	role := &models.Role{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Level:       req.Level,
		Permissions: req.Permissions,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	return s.roleRepo.CreateRole(ctx, role)
}

func (s *RoleService) GetRoles() ([]*models.Role, error) {
	return s.roleRepo.GetRole("")
}

func (s *RoleService) GetRoleByID(id string) (*models.Role, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("role ID is required")
	}

	roles, err := s.roleRepo.GetRole(id)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, errors.New("role not found")
	}

	return roles[0], nil
}

func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("role name is required")
	}

	roles, err := s.roleRepo.GetRole(name)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, errors.New("role not found")
	}

	return roles[0], nil
}

func (s *RoleService) UpdateRole(id string, req *models.UpdateRoleRequest, updatedBy string) (*models.Role, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("role ID is required")
	}

	if err := s.validateUpdateRoleRequest(req); err != nil {
		return nil, err
	}

	role := &models.Role{
		UpdatedBy: updatedBy,
	}

	if req.Name != "" {
		role.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		role.Description = strings.TrimSpace(req.Description)
	}
	if req.Level != nil {
		role.Level = *req.Level
	}
	if req.Permissions != nil {
		role.Permissions = req.Permissions
	}
	if req.Status != "" {
		role.Status = req.Status
	}

	return s.roleRepo.UpdateRole(id, role)
}

func (s *RoleService) DeleteRole(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("role ID is required")
	}

	return s.roleRepo.DeleteRole(id)
}

func (s *RoleService) GetRolesByStatus(status models.RoleStatus) ([]*models.Role, error) {
	if status == "" {
		return nil, errors.New("status is required")
	}

	return s.roleRepo.GetRolesByStatus(status)
}

func (s *RoleService) validateCreateRoleRequest(req *models.CreateRoleRequest) error {
	if req == nil {
		return errors.New("create role request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.New("role name is required")
	}

	if len(req.Name) > 100 {
		return errors.New("role name must be less than 100 characters")
	}

	if strings.TrimSpace(req.Description) == "" {
		return errors.New("role description is required")
	}

	if len(req.Description) > 500 {
		return errors.New("role description must be less than 500 characters")
	}

	if req.Level < 1 || req.Level > 10 {
		return errors.New("role level must be between 1 and 10")
	}

	if len(req.Permissions) == 0 {
		return errors.New("at least one permission is required")
	}

	for _, permission := range req.Permissions {
		if strings.TrimSpace(permission) == "" {
			return errors.New("permission cannot be empty")
		}
	}

	return nil
}

func (s *RoleService) validateUpdateRoleRequest(req *models.UpdateRoleRequest) error {
	if req == nil {
		return errors.New("update role request is required")
	}

	if req.Name != "" && len(req.Name) > 100 {
		return errors.New("role name must be less than 100 characters")
	}

	if req.Description != "" && len(req.Description) > 500 {
		return errors.New("role description must be less than 500 characters")
	}

	if req.Level != nil && (*req.Level < 1 || *req.Level > 10) {
		return errors.New("role level must be between 1 and 10")
	}

	if req.Status != "" {
		validStatuses := []models.RoleStatus{
			models.RoleStatusActive,
			models.RoleStatusInactive,
			models.RoleStatusArchived,
		}
		isValid := false
		for _, status := range validStatuses {
			if req.Status == status {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid status: %s", req.Status)
		}
	}

	if req.Permissions != nil {
		for _, permission := range req.Permissions {
			if strings.TrimSpace(permission) == "" {
				return errors.New("permission cannot be empty")
			}
		}
	}

	return nil
}
