package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"strings"
	"time"
)

type JobService struct {
	jobRepo repository.JobRepositoryInterface
	logger  logger.Logger
}

func NewJobService(jobRepo repository.JobRepositoryInterface, logger logger.Logger) *JobService {
	return &JobService{
		jobRepo: jobRepo,
		logger:  logger,
	}
}

func (s *JobService) CreateJob(ctx context.Context, req *models.CreateJobRequest, createdBy string) (*models.Job, error) {
	if err := s.validateCreateJob(req); err != nil {
		return nil, err
	}

	job := &models.Job{
		ClientID:              req.ClientID,
		JobsName:              req.JobsName,
		JobType:               req.JobType,
		Notes:                 req.Notes,
		OrgID:                 req.OrgID,
		UsersAssignedToJob:    req.UsersAssignedToJob,
		VehiclesAssignedToJob: req.VehiclesAssignedToJob,
		QBInfoOnJob:           req.QBInfoOnJob,
		CreatedData: models.CreatedData{
			UID:        createdBy,
			UserStatus: "active",
		},
		JobImagesAfterService: []string{},
		JobStatus:             models.JobStatusPending,
	}

	return s.jobRepo.CreateJob(ctx, job)
}

func (s *JobService) validateCreateJob(req *models.CreateJobRequest) error {
	if req == nil {
		return errors.New("job request is required")
	}

	if strings.TrimSpace(req.JobsName) == "" {
		return errors.New("job name is required")
	}

	if len(req.JobsName) < 2 || len(req.JobsName) > 200 {
		return errors.New("job name must be between 2 and 200 characters")
	}

	if strings.TrimSpace(req.ClientID) == "" {
		return errors.New("client ID is required")
	}

	if strings.TrimSpace(req.OrgID) == "" {
		return errors.New("organization ID is required")
	}

	if req.JobType == "" {
		return errors.New("job type is required")
	}

	if len(req.Notes) > 1000 {
		return errors.New("notes must be less than 1000 characters")
	}

	return nil
}

func (s *JobService) GetJobs(filter *models.JobFilter) ([]*models.Job, error) {
	if filter == nil {
		filter = &models.JobFilter{}
	}
	return s.jobRepo.GetJobsByFilter(filter)
}

func (s *JobService) GetJobByID(id string) (*models.Job, error) {
	jobs, err := s.jobRepo.GetJob(id)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, errors.New("job not found")
	}
	return jobs[0], nil
}

func (s *JobService) UpdateJob(ctx context.Context, id string, req *models.UpdateJobRequest, updatedBy string) (*models.Job, error) {
	if err := s.validateUpdateJob(req); err != nil {
		return nil, err
	}

	// Get existing job
	existing, err := s.GetJobByID(id)
	if err != nil {
		return nil, err
	}

	// Apply updates to existing job
	updatedJob := *existing
	updatedJob.UpdatedBy = updatedBy

	if req.JobsName != "" {
		updatedJob.JobsName = req.JobsName
	}
	if req.JobStatus != "" {
		updatedJob.JobStatus = req.JobStatus
	}
	if req.JobType != "" {
		updatedJob.JobType = req.JobType
	}
	if req.Notes != "" {
		updatedJob.Notes = req.Notes
	}
	if req.UsersAssignedToJob != nil {
		updatedJob.UsersAssignedToJob = req.UsersAssignedToJob
	}
	if req.VehiclesAssignedToJob != nil {
		updatedJob.VehiclesAssignedToJob = req.VehiclesAssignedToJob
	}
	if req.QBInfoOnJob != nil {
		updatedJob.QBInfoOnJob = req.QBInfoOnJob
	}
	if req.JobImagesAfterService != nil {
		updatedJob.JobImagesAfterService = req.JobImagesAfterService
	}

	return s.jobRepo.UpdateJob(id, &updatedJob)
}

func (s *JobService) validateUpdateJob(req *models.UpdateJobRequest) error {
	if req == nil {
		return errors.New("update request is required")
	}

	if req.JobsName != "" && (len(req.JobsName) < 2 || len(req.JobsName) > 200) {
		return errors.New("job name must be between 2 and 200 characters")
	}

	if len(req.Notes) > 1000 {
		return errors.New("notes must be less than 1000 characters")
	}

	return nil
}

func (s *JobService) DeleteJob(id string) error {
	return s.jobRepo.DeleteJob(id)
}

func (s *JobService) StartJob(ctx context.Context, id string, startedBy string) (*models.Job, error) {
	existing, err := s.GetJobByID(id)
	if err != nil {
		return nil, err
	}

	if existing.JobStatus != models.JobStatusPending && existing.JobStatus != models.JobStatusActive {
		return nil, errors.New("job cannot be started from current status")
	}

	now := time.Now()
	updatedJob := *existing
	updatedJob.JobStatus = models.JobStatusInProgress
	updatedJob.JobStartedAt = &now
	updatedJob.StartedData = &models.StartedData{
		UID:        startedBy,
		UserStatus: "active",
		StartedAt:  now,
	}

	return s.jobRepo.UpdateJob(id, &updatedJob)
}

func (s *JobService) CompleteJob(ctx context.Context, id string, completedBy string) (*models.Job, error) {
	existing, err := s.GetJobByID(id)
	if err != nil {
		return nil, err
	}

	if existing.JobStatus != models.JobStatusInProgress {
		return nil, errors.New("job must be in progress to be completed")
	}

	now := time.Now()
	updatedJob := *existing
	updatedJob.JobStatus = models.JobStatusCompleted
	updatedJob.JobEndedAt = &now
	updatedJob.UpdatedBy = completedBy

	return s.jobRepo.UpdateJob(id, &updatedJob)
}

func (s *JobService) CancelJob(ctx context.Context, id string, cancelledBy string, reason string) (*models.Job, error) {
	existing, err := s.GetJobByID(id)
	if err != nil {
		return nil, err
	}

	if existing.JobStatus == models.JobStatusCompleted {
		return nil, errors.New("completed job cannot be cancelled")
	}

	now := time.Now()
	updatedJob := *existing
	updatedJob.JobStatus = models.JobStatusCancelled
	updatedJob.DeletedData = &models.DeletedData{
		UID:        cancelledBy,
		UserStatus: "active",
		DeletedAt:  now,
		Reason:     reason,
	}
	updatedJob.UpdatedBy = cancelledBy

	return s.jobRepo.UpdateJob(id, &updatedJob)
}

func (s *JobService) GetJobsByOrganization(orgID string, status models.JobStatus) ([]*models.Job, error) {
	filter := &models.JobFilter{
		OrgID: orgID,
	}
	if status != "" {
		filter.JobStatus = status
	}
	return s.jobRepo.GetJobsByFilter(filter)
}

func (s *JobService) GetJobsByClient(clientID string) ([]*models.Job, error) {
	filter := &models.JobFilter{
		ClientID: clientID,
	}
	return s.jobRepo.GetJobsByFilter(filter)
}