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
	"strconv"
	"strings"
	"time"
)

type CrewRepository struct {
	db     dal.DatabaseClientInterface
	config *models.Config
	logger logger.Logger
}

func NewCrewRepository(db dal.DatabaseClientInterface, cfg *models.Config, log logger.Logger) *CrewRepository {
	return &CrewRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *CrewRepository) CreateCrew(ctx context.Context, crew *models.Crew) (*models.Crew, error) {
	r.logger.Infof("Creating crew: %s", crew.Name)

	now := time.Now()
	crew.CrewID = "crew_" + utils.GenerateUUID()
	crew.CreatedAt = now
	crew.UpdatedAt = now
	crew.IsActive = true

	err := r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_crews", crew)
	if err != nil {
		r.logger.Errorf("Failed to create crew: %v", err)
		return nil, err
	}

	r.logger.Infof("Crew created successfully: %s", crew.CrewID)
	return crew, nil
}

func (r *CrewRepository) GetCrew(key string) ([]*models.Crew, error) {
	ctx := context.Background()

	if key == "" {
		return nil, errors.New("crew key is required")
	}

	r.logger.Infof("Crew checking for: %s", key)

	crew := models.Crew{}
	keyType, indexName, keyName := r.determineKeyType(key)

	var config models.QueryConfig

	if keyType == "id" {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_crews",
			KeyName:   "crewID",
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	} else {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_crews",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	}

	r.logger.Infof("Querying %s table with %s: %s", r.config.DynamoDBTablePrefix, keyName, key)

	err := r.db.GetItem(ctx, config, &crew)
	if err != nil {
		r.logger.Errorf("Failed to get crew by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get crew by %s: %w", keyName, err)
	}

	if crew.CrewID == "" {
		return nil, errors.New("crew not found")
	}

	r.logger.Infof("Crew found: %s", crew.CrewID)
	return []*models.Crew{&crew}, nil
}

func (r *CrewRepository) GetCrewsByFilter(filter *models.CrewFilter) ([]*models.Crew, error) {
	ctx := context.Background()
	r.logger.Infof("Getting crews with filter")

	var crews []*models.Crew
	var err error

	if filter.OrgID != "" {
		err = r.db.QueryByIndex(ctx,
			r.config.DynamoDBTablePrefix+"_crews",
			"orgID-index",
			"orgID", filter.OrgID,
			&crews)
	} else if filter.LeadTechnicianId != "" {
		err = r.db.QueryByIndex(ctx,
			r.config.DynamoDBTablePrefix+"_crews",
			"leadTechnicianId-index",
			"leadTechnicianId", filter.LeadTechnicianId,
			&crews)
	} else if filter.CreatedBy != "" {
		err = r.db.QueryByIndex(ctx,
			r.config.DynamoDBTablePrefix+"_crews",
			"createdBy-index",
			"createdBy", filter.CreatedBy,
			&crews)
	} else if filter.IsActive != nil {
		activeStr := strconv.FormatBool(*filter.IsActive)
		err = r.db.QueryByIndex(ctx,
			r.config.DynamoDBTablePrefix+"_crews",
			"isActive-index",
			"isActive", activeStr,
			&crews)
	} else {
		err = r.db.ScanTable(ctx, r.config.DynamoDBTablePrefix+"_crews", &crews)
	}

	if err != nil {
		r.logger.Errorf("Failed to get crews: %v", err)
		return nil, err
	}

	filteredCrews := r.applyAdditionalFilters(crews, filter)

	r.logger.Infof("Found %d crews", len(filteredCrews))
	return filteredCrews, nil
}

func (r *CrewRepository) UpdateCrew(id string, crew *models.Crew) (*models.Crew, error) {
	ctx := context.Background()
	r.logger.Infof("Updating crew: %s", id)

	if id == "" {
		return nil, errors.New("crew ID is required")
	}

	existing, err := r.GetCrew(id)
	if err != nil {
		return nil, fmt.Errorf("crew not found: %w", err)
	}
	if len(existing) == 0 {
		return nil, errors.New("crew not found")
	}

	now := time.Now()
	crew.CrewID = id
	crew.CreatedAt = existing[0].CreatedAt
	crew.CreatedBy = existing[0].CreatedBy
	crew.UpdatedAt = now

	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_crews", crew)
	if err != nil {
		r.logger.Errorf("Failed to update crew: %v", err)
		return nil, err
	}

	r.logger.Infof("Crew updated successfully: %s", id)
	return crew, nil
}

func (r *CrewRepository) DeleteCrew(id string) error {
	ctx := context.Background()
	r.logger.Infof("Deleting crew: %s", id)

	if id == "" {
		return errors.New("crew ID is required")
	}

	err := r.db.DeleteItem(ctx, r.config.DynamoDBTablePrefix+"_crews", "crewID", id)
	if err != nil {
		r.logger.Errorf("Failed to delete crew: %v", err)
		return err
	}

	r.logger.Infof("Crew deleted successfully: %s", id)
	return nil
}

func (r *CrewRepository) determineKeyType(key string) (keyType, indexName, keyName string) {
	crewIDPattern := `^crew_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	
	isCrewID, _ := regexp.MatchString(crewIDPattern, strings.ToLower(key))
	isUUID, _ := regexp.MatchString(uuidPattern, strings.ToLower(key))

	if isCrewID {
		return "id", "", "crewID"
	} else if isUUID {
		return "leadTechnician", "leadTechnicianId-index", "leadTechnicianId"
	} else {
		return "orgID", "orgID-index", "orgID"
	}
}

func (r *CrewRepository) applyAdditionalFilters(crews []*models.Crew, filter *models.CrewFilter) []*models.Crew {
	if filter == nil {
		return crews
	}

	var filtered []*models.Crew
	for _, crew := range crews {
		if filter.CreatedBy != "" && crew.CreatedBy != filter.CreatedBy {
			continue
		}

		if !filter.FromDate.IsZero() && crew.CreatedAt.Before(filter.FromDate) {
			continue
		}
		if !filter.ToDate.IsZero() && crew.CreatedAt.After(filter.ToDate) {
			continue
		}

		if filter.IsActive != nil && crew.IsActive != *filter.IsActive {
			continue
		}

		if filter.LeadTechnicianId != "" && crew.LeadTechnicianId != filter.LeadTechnicianId {
			continue
		}

		filtered = append(filtered, crew)
	}

	return filtered
}