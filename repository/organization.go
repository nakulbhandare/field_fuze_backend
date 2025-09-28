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

// OrganizationRepository implements OrganizationRepositoryInterface
type OrganizationRepository struct {
	db     dal.DatabaseClientInterface
	config *models.Config
	logger logger.Logger
}

func NewOrganizationRepository(db dal.DatabaseClientInterface, cfg *models.Config, log logger.Logger) *OrganizationRepository {
	return &OrganizationRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *OrganizationRepository) CreateOrganization(ctx context.Context, organization *models.Organization) (*models.Organization, error) {
	r.logger.Infof("Creating organization: %s", organization.Name)

	existingOrg := &models.Organization{}
	err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_organization", "name-index", "name", organization.Name, &[]*models.Organization{existingOrg})
	if err == nil && existingOrg.ID != "" {
		return nil, errors.New("organization with this name already exists")
	}

	now := time.Now()
	organization.ID = utils.GenerateUUID()
	organization.CreatedAt = now

	fmt.Println("organization ::::", dal.PrintPrettyJSON(organization))

	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_organization", organization)
	if err != nil {
		r.logger.Errorf("Failed to create organization: %v", err)
		return nil, err
	}

	r.logger.Infof("Organization created successfully: %s", organization.ID)
	return organization, nil
}

func (r *OrganizationRepository) GetOrganization(key string) ([]*models.Organization, error) {
	ctx := context.Background()

	if key == "" {
		return nil, errors.New("organization ID is required")
	}

	r.logger.Infof("Organization checking for: %s", key)

	organization := models.Organization{}
	keyType, indexName, keyName := r.determineKeyType(key)

	var config models.QueryConfig

	if keyType == "id" {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			KeyName:   "id",
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	} else {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	}

	r.logger.Infof("Querying %s table with %s: %s", r.config.DynamoDBTablePrefix, keyName, key)

	err := r.db.GetItem(ctx, config, &organization)
	if err != nil {
		r.logger.Errorf("Failed to get organization by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get organization by %s: %w", keyName, err)
	}

	if organization.ID == "" {
		return nil, errors.New("organization not found")
	}

	r.logger.Infof("Organization found: %s", organization.ID)
	return []*models.Organization{&organization}, nil
}

func (r *OrganizationRepository) determineKeyType(key string) (keyType, indexName, keyName string) {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	isUUID, _ := regexp.MatchString(uuidPattern, strings.ToLower(key))

	if isUUID {
		return "id", "", "id"
	} else {
		return "name", "name-index", "name"
	}
}
