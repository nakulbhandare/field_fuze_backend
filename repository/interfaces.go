package repository

import (
	"context"
	"fieldfuze-backend/models"
)

// UserRepositoryInterface defines the contract for user repository operations
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(key string) ([]*models.User, error)
	UpdateUser(id string, user *models.User) (*models.User, error)
	AssignRoles(ctx context.Context, userID string, roleAssignments []models.RoleAssignment) (*models.User, error)
	AddRoleToUser(ctx context.Context, userID string, roleAssignment models.RoleAssignment) (*models.User, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) (*models.User, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) (*models.User, error)
}

// RoleRepositoryInterface defines the contract for role repository operations
type RoleRepositoryInterface interface {
	CreateRoleAssignment(ctx context.Context, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error)
	GetRoleAssignments(id string) ([]*models.RoleAssignment, error)
	GetRole(name string) ([]*models.Role, error)
	UpdateRole(id string, role *models.Role) (*models.Role, error)
	DeleteRole(id string) error
	GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error)
	UpdateRoleAssignment(id string, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error)
	DeleteRoleAssignment(id string) error
}

// RepositoryContainerInterface defines the contract for the repository container
type RepositoryContainerInterface interface {
	GetUserRepository() UserRepositoryInterface
	GetRoleRepository() RoleRepositoryInterface
	GetOrganizationRepository() OrganizationRepositoryInterface
	GetJobRepository() JobRepositoryInterface
}

// OrganizationRepositoryInterface defines the contract for the organization repository
type OrganizationRepositoryInterface interface {
	CreateOrganization(ctx context.Context, organization *models.Organization) (*models.Organization, error)
	GetOrganization(name string) ([]*models.Organization, error)
	UpdateOrganization(id string, organization *models.Organization) (*models.Organization, error)
	DeleteOrganization(id string) error
}

// JobRepositoryInterface defines the contract for job repository operations
type JobRepositoryInterface interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	GetJob(key string) ([]*models.Job, error)
	GetJobsByFilter(filter *models.JobFilter) ([]*models.Job, error)
	UpdateJob(id string, job *models.Job) (*models.Job, error)
	DeleteJob(id string) error
}
