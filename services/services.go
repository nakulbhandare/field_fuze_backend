package services

import (
	"context"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
)

// Service implements ServiceContainerInterface
type Service struct {
	userService           UserServiceInterface
	roleService           RoleServiceInterface
	infrastructureService InfrastructureServiceInterface
	organizationService   OrganizationServiceInterface
	jobService            JobServiceInterface
}

// NewService creates a new service container with all dependencies injected
func NewService(
	ctx context.Context,
	repoContainer repository.RepositoryContainerInterface,
	dalContainer dal.DALContainerInterface,
	logger logger.Logger,
	config *models.Config,
) ServiceContainerInterface {
	return &Service{
		userService:           NewUserService(ctx, repoContainer.GetUserRepository(), logger),
		roleService:           NewRoleService(repoContainer.GetRoleRepository(), logger),
		infrastructureService: NewInfrastructureService(ctx, dalContainer.GetDatabaseClient(), logger, config),
		organizationService:   NewOrganizationService(repoContainer.GetOrganizationRepository(), logger),
		jobService:            NewJobService(repoContainer.GetJobRepository(), logger),
	}
}

// GetUserService returns the user service interface
func (s *Service) GetUserService() UserServiceInterface {
	return s.userService
}

// GetRoleService returns the role service interface
func (s *Service) GetRoleService() RoleServiceInterface {
	return s.roleService
}

// GetInfrastructureService returns the infrastructure service interface
func (s *Service) GetInfrastructureService() InfrastructureServiceInterface {
	return s.infrastructureService
}

// GetOrganizationService returns the organization service interface
func (s *Service) GetOrganizationService() OrganizationServiceInterface {
	return s.organizationService
}

// GetJobService returns the job service interface
func (s *Service) GetJobService() JobServiceInterface {
	return s.jobService
}
