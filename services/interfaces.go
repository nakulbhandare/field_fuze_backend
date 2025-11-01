package services

import (
	"context"
	"fieldfuze-backend/models"
)

// UserServiceInterface defines the contract for user service
type UserServiceInterface interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUsers() ([]*models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(id string, user *models.User) (*models.User, error)
	AssignRolesToUser(userID string, roleAssignments []models.RoleAssignment) (*models.User, error)
	AddRoleToUser(userID string, roleAssignment models.RoleAssignment) (*models.User, error)
	AssignRoleToUser(userID, roleID string) (*models.User, error)
	RemoveRoleFromUser(userID, roleID string) (*models.User, error)
	GetUsersByStatus(status models.UserStatus) ([]*models.User, error)
}

// RoleServiceInterface defines the contract for role service
type RoleServiceInterface interface {
	CreateRole(ctx context.Context, roleAssignment *models.RoleAssignment, createdBy string) (*models.RoleAssignment, error)
	GetRoleAssignments() ([]*models.RoleAssignment, error)
	GetRoleAssignmentByID(id string) (*models.RoleAssignment, error)
	GetRoleByName(name string) (*models.Role, error)
	UpdateRole(id string, req *models.UpdateRoleRequest, updatedBy string) (*models.Role, error)
	DeleteRole(id string) error
	GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error)
	UpdateRoleAssignment(id string, roleAssignment *models.RoleAssignment, updatedBy string) (*models.RoleAssignment, error)
	DeleteRoleAssignment(id string) error
}

// InfrastructureServiceInterface defines the contract for infrastructure service
type InfrastructureServiceInterface interface {
	GetWorkerStatus(ctx context.Context) (*models.ExecutionResult, error)
	RestartWorker(ctx context.Context, force bool) (*models.ServiceRestartResult, error)
	IsWorkerHealthy() (bool, string, error)
	AutoRestartIfNeeded(ctx context.Context) (*models.ServiceRestartResult, error)
}

// OrganizationServiceInterface defines the contract for organization service
type OrganizationServiceInterface interface {
	CreateOrganization(ctx context.Context, organization *models.Organization, createdBy string) (*models.Organization, error)
	GetOrganizations(key string) ([]*models.Organization, error)
	GetOrganizationByID(id string) (*models.Organization, error)
	UpdateOrganization(id string, req *models.Organization, updatedBy string) (*models.Organization, error)
	DeleteOrganization(id string) error
	GetOrganizationAssignmentsByStatus(status string) ([]*models.Organization, error)
	UpdateOrganizationAssignment(id string, organizationAssignment *models.Organization, updatedBy string) (*models.Organization, error)
	DeleteOrganizationAssignment(id string) error
}

// JobServiceInterface defines the contract for job service
type JobServiceInterface interface {
	CreateJob(ctx context.Context, req *models.CreateJobRequest, createdBy string) (*models.Job, error)
	GetJobs(filter *models.JobFilter) ([]*models.Job, error)
	GetJobByID(id string) (*models.Job, error)
	UpdateJob(ctx context.Context, id string, req *models.UpdateJobRequest, updatedBy string) (*models.Job, error)
	DeleteJob(id string) error
	StartJob(ctx context.Context, id string, startedBy string) (*models.Job, error)
	CompleteJob(ctx context.Context, id string, completedBy string) (*models.Job, error)
	CancelJob(ctx context.Context, id string, cancelledBy string, reason string) (*models.Job, error)
	GetJobsByOrganization(orgID string, status models.JobStatus) ([]*models.Job, error)
	GetJobsByClient(clientID string) ([]*models.Job, error)
}

// CrewServiceInterface defines the contract for crew service
type CrewServiceInterface interface {
	CreateCrew(ctx context.Context, req *models.CreateCrewRequest, createdBy string) (*models.Crew, error)
	GetCrews(filter *models.CrewFilter) ([]*models.Crew, error)
	GetCrewByID(id string) (*models.Crew, error)
	UpdateCrew(ctx context.Context, id string, req *models.UpdateCrewRequest) (*models.Crew, error)
	DeleteCrew(id string) error
	GetCrewsByOrganization(orgID string, isActive *bool) ([]*models.Crew, error)
	GetCrewsByLeadTechnician(leadTechnicianId string) ([]*models.Crew, error)
	AddMemberToCrew(ctx context.Context, crewID, memberID string) (*models.Crew, error)
	RemoveMemberFromCrew(ctx context.Context, crewID, memberID string) (*models.Crew, error)
}

// ServiceContainer interface defines the main service container contract
type ServiceContainerInterface interface {
	GetUserService() UserServiceInterface
	GetRoleService() RoleServiceInterface
	GetInfrastructureService() InfrastructureServiceInterface
	GetOrganizationService() OrganizationServiceInterface
	GetJobService() JobServiceInterface
	GetCrewService() CrewServiceInterface
}
