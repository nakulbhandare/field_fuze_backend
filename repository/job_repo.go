package repository

import (
	"context"
	"errors"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type JobRepository struct {
	db     dal.DatabaseClientInterface
	config *models.Config
	logger logger.Logger
}

func NewJobRepository(db dal.DatabaseClientInterface, cfg *models.Config, log logger.Logger) *JobRepository {
	return &JobRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *JobRepository) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	r.logger.Infof("Creating job: %s", job.JobsName)

	now := time.Now()
	job.JobID = utils.GenerateUUID()
	job.CreatedAt = now
	job.UpdatedAt = now
	job.JobStatus = models.JobStatusPending

	fmt.Println("job ::::", dal.PrintPrettyJSON(job))

	err := r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_jobs", job)
	if err != nil {
		r.logger.Errorf("Failed to create job: %v", err)
		return nil, err
	}

	r.logger.Infof("Job created successfully: %s", job.JobID)
	return job, nil
}

func (r *JobRepository) GetJob(key string) ([]*models.Job, error) {
	ctx := context.Background()

	if key == "" {
		return nil, errors.New("job key is required")
	}

	r.logger.Infof("Job checking for: %s", key)

	job := models.Job{}
	keyType, indexName, keyName := r.determineKeyType(key)

	var config models.QueryConfig

	if keyType == "id" {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_jobs",
			KeyName:   "jobID",
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	} else {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_jobs",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	}

	r.logger.Infof("Querying %s table with %s: %s", r.config.DynamoDBTablePrefix, keyName, key)

	err := r.db.GetItem(ctx, config, &job)
	if err != nil {
		r.logger.Errorf("Failed to get job by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get job by %s: %w", keyName, err)
	}

	if job.JobID == "" {
		return nil, errors.New("job not found")
	}

	r.logger.Infof("Job found: %s", job.JobID)
	return []*models.Job{&job}, nil
}

func (r *JobRepository) GetJobsByFilter(filter *models.JobFilter) ([]*models.Job, error) {
	ctx := context.Background()
	r.logger.Infof("Getting jobs with filter")

	var jobs []*models.Job
	var err error

	// Query by organization 
	if filter.OrgID != "" {
		// Use simple orgID index (date filtering will be done post-query)
		err = r.db.QueryByIndex(ctx, 
			r.config.DynamoDBTablePrefix+"_jobs", 
			"orgID-index", 
			"orgID", filter.OrgID, 
			&jobs)
	} else if filter.ClientID != "" {
		err = r.db.QueryByIndex(ctx, 
			r.config.DynamoDBTablePrefix+"_jobs", 
			"clientID-index", 
			"clientID", filter.ClientID, 
			&jobs)
	} else if filter.JobStatus != "" {
		err = r.db.QueryByIndex(ctx, 
			r.config.DynamoDBTablePrefix+"_jobs", 
			"jobStatus-index", 
			"jobStatus", string(filter.JobStatus), 
			&jobs)
	} else if filter.JobType != "" {
		err = r.db.QueryByIndex(ctx, 
			r.config.DynamoDBTablePrefix+"_jobs", 
			"jobType-index", 
			"jobType", string(filter.JobType), 
			&jobs)
	} else {
		// Scan all jobs (use with caution in production)
		err = r.db.ScanTable(ctx, r.config.DynamoDBTablePrefix+"_jobs", &jobs)
	}

	if err != nil {
		r.logger.Errorf("Failed to get jobs: %v", err)
		return nil, err
	}

	// Apply additional filtering if needed
	filteredJobs := r.applyAdditionalFilters(jobs, filter)

	r.logger.Infof("Found %d jobs", len(filteredJobs))
	return filteredJobs, nil
}

func (r *JobRepository) UpdateJob(id string, job *models.Job) (*models.Job, error) {
	ctx := context.Background()
	r.logger.Infof("Updating job: %s", id)

	if id == "" {
		return nil, errors.New("job ID is required")
	}

	existing, err := r.GetJob(id)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	if len(existing) == 0 {
		return nil, errors.New("job not found")
	}

	now := time.Now()
	job.JobID = id
	job.CreatedAt = existing[0].CreatedAt
	job.UpdatedAt = now

	// Handle job state transitions
	if job.JobStatus == models.JobStatusInProgress && existing[0].JobStatus != models.JobStatusInProgress {
		job.JobStartedAt = &now
	}
	if job.JobStatus == models.JobStatusCompleted && existing[0].JobStatus != models.JobStatusCompleted {
		job.JobEndedAt = &now
	}

	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_jobs", job)
	if err != nil {
		r.logger.Errorf("Failed to update job: %v", err)
		return nil, err
	}

	r.logger.Infof("Job updated successfully: %s", id)
	return job, nil
}

func (r *JobRepository) DeleteJob(id string) error {
	ctx := context.Background()
	r.logger.Infof("Deleting job: %s", id)

	if id == "" {
		return errors.New("job ID is required")
	}

	err := r.db.DeleteItem(ctx, r.config.DynamoDBTablePrefix+"_jobs", "jobID", id)
	if err != nil {
		r.logger.Errorf("Failed to delete job: %v", err)
		return err
	}

	r.logger.Infof("Job deleted successfully: %s", id)
	return nil
}

func (r *JobRepository) determineKeyType(key string) (keyType, indexName, keyName string) {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	isUUID, _ := regexp.MatchString(uuidPattern, strings.ToLower(key))

	if isUUID {
		return "id", "", "jobID"
	} else {
		// Assume it's a client ID or organization ID
		return "clientID", "clientID-index", "clientID"
	}
}

func (r *JobRepository) applyAdditionalFilters(jobs []*models.Job, filter *models.JobFilter) []*models.Job {
	if filter == nil {
		return jobs
	}

	var filtered []*models.Job
	for _, job := range jobs {
		// Apply CreatedBy filter
		if filter.CreatedBy != "" && job.CreatedData.UID != filter.CreatedBy {
			continue
		}

		// Apply date range filter if not already applied in query
		if !filter.FromDate.IsZero() && job.CreatedAt.Before(filter.FromDate) {
			continue
		}
		if !filter.ToDate.IsZero() && job.CreatedAt.After(filter.ToDate) {
			continue
		}

		// Apply JobStatus filter if not already applied in query
		if filter.JobStatus != "" && job.JobStatus != filter.JobStatus {
			continue
		}

		// Apply JobType filter if not already applied in query
		if filter.JobType != "" && job.JobType != filter.JobType {
			continue
		}

		filtered = append(filtered, job)
	}

	return filtered
}