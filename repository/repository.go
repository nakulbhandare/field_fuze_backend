package repository

import (
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils/logger"
)

// Repository implements RepositoryContainerInterface
type Repository struct {
	userRepository         UserRepositoryInterface
	roleRepository         RoleRepositoryInterface
	organizationRepository OrganizationRepositoryInterface
	jobRepository          JobRepositoryInterface
}

// NewRepository creates a new repository container with all dependencies injected
func NewRepository(dalContainer dal.DALContainerInterface, cfg *models.Config, log logger.Logger) RepositoryContainerInterface {
	dbClient := dalContainer.GetDatabaseClient()
	
	return &Repository{
		userRepository:         NewUserRepository(dbClient, cfg, log),
		roleRepository:         NewRoleRepository(dbClient, cfg, log),
		organizationRepository: NewOrganizationRepository(dbClient, cfg, log),
		jobRepository:          NewJobRepository(dbClient, cfg, log),
	}
}

// GetUserRepository returns the user repository interface
func (r *Repository) GetUserRepository() UserRepositoryInterface {
	return r.userRepository
}

// GetRoleRepository returns the role repository interface
func (r *Repository) GetRoleRepository() RoleRepositoryInterface {
	return r.roleRepository
}

// GetOrganizationRepository returns the organization repository interface
func (r *Repository) GetOrganizationRepository() OrganizationRepositoryInterface {
	return r.organizationRepository
}

// GetJobRepository returns the job repository interface
func (r *Repository) GetJobRepository() JobRepositoryInterface {
	return r.jobRepository
}
