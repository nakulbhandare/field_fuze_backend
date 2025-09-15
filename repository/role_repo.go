package repository

import (
	"context"
	"errors"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils"
	"fmt"
	"regexp"
	"strings"
	"time"

	"fieldfuze-backend/utils/logger"
)

type RoleRepository struct {
	db     *dal.DynamoDBClient
	config *models.Config
	logger logger.Logger
}

func NewRoleRepository(db *dal.DynamoDBClient, cfg *models.Config, log logger.Logger) *RoleRepository {
	return &RoleRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *RoleRepository) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	r.logger.Infof("Creating role: %s", role.Name)

	existingRole := &models.Role{}
	err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_roles", "name-index", "name", role.Name, &[]*models.Role{existingRole})
	if err == nil && existingRole.ID != "" {
		return nil, errors.New("role with this name already exists")
	}

	now := time.Now()
	role.ID = utils.GenerateUUID()
	role.CreatedAt = now
	role.UpdatedAt = now
	role.Status = models.RoleStatusActive

	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_roles", role)
	if err != nil {
		r.logger.Errorf("Failed to create role: %v", err)
		return nil, err
	}

	r.logger.Infof("Role created successfully: %s", role.ID)
	return role, nil
}

func (r *RoleRepository) GetRole(key string) ([]*models.Role, error) {
	ctx := context.Background()

	if key == "" {
		var roles []*models.Role
		tableName := r.config.DynamoDBTablePrefix + "_roles"

		r.logger.Infof("Scanning %s table for all roles", tableName)

		err := r.db.ScanTable(ctx, tableName, &roles)
		if err != nil {
			r.logger.Errorf("Failed to scan roles table: %v", err)
			return nil, fmt.Errorf("failed to get all roles: %w", err)
		}

		r.logger.Infof("Found %d roles", len(roles))
		return roles, nil
	}

	role := models.Role{}
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

	err := r.db.GetItem(ctx, config, &role)
	if err != nil {
		r.logger.Errorf("Failed to get role by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get role by %s: %w", keyName, err)
	}

	if role.ID == "" {
		return nil, errors.New("role not found")
	}

	r.logger.Infof("Role found: %s", role.ID)
	return []*models.Role{&role}, nil
}

func (r *RoleRepository) determineKeyType(key string) (keyType, indexName, keyName string) {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	isUUID, _ := regexp.MatchString(uuidPattern, strings.ToLower(key))

	if isUUID {
		return "id", "", "id"
	} else {
		return "name", "name-index", "name"
	}
}

func (r *RoleRepository) UpdateRole(id string, role *models.Role) (*models.Role, error) {
	ctx := context.Background()

	existingRole := models.Role{}
	keyType, indexName, keyName := r.determineKeyType(id)

	var config models.QueryConfig

	if keyType == "id" {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			KeyName:   "id",
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	} else {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	}

	err := r.db.GetItem(ctx, config, &existingRole)
	if err != nil {
		r.logger.Errorf("Failed to get role by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get role by %s: %w", keyName, err)
	}

	if existingRole.ID == "" {
		return nil, errors.New("role not found")
	}

	updates := make(map[string]interface{})

	if role.Name != "" {
		existingRoleWithName := &models.Role{}
		err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_roles", "name-index", "name", role.Name, &[]*models.Role{existingRoleWithName})
		if err == nil && existingRoleWithName.ID != "" && existingRoleWithName.ID != existingRole.ID {
			return nil, errors.New("role with this name already exists")
		}
		updates["name"] = role.Name
	}
	if role.Description != "" {
		updates["description"] = role.Description
	}
	if role.Level != 0 {
		updates["level"] = role.Level
	}
	if role.Permissions != nil {
		updates["permissions"] = role.Permissions
	}
	if role.Status != "" {
		updates["status"] = role.Status
	}
	if role.UpdatedBy != "" {
		updates["updated_by"] = role.UpdatedBy
	}
	if role.Metadata != nil {
		updates["metadata"] = role.Metadata
	}
	updates["updated_at"] = time.Now()

	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_roles", "id", existingRole.ID, updates)
	if err != nil {
		r.logger.Errorf("Failed to update role: %v", err)
		return nil, err
	}

	if role.Name != "" {
		existingRole.Name = role.Name
	}
	if role.Description != "" {
		existingRole.Description = role.Description
	}
	if role.Level != 0 {
		existingRole.Level = role.Level
	}
	if role.Permissions != nil {
		existingRole.Permissions = role.Permissions
	}
	if role.Status != "" {
		existingRole.Status = role.Status
	}
	if role.UpdatedBy != "" {
		existingRole.UpdatedBy = role.UpdatedBy
	}
	if role.Metadata != nil {
		existingRole.Metadata = role.Metadata
	}
	existingRole.UpdatedAt = time.Now()

	r.logger.Infof("Role updated successfully: %s", existingRole.ID)
	return &existingRole, nil
}

func (r *RoleRepository) DeleteRole(id string) error {
	ctx := context.Background()

	existingRole := models.Role{}
	keyType, indexName, keyName := r.determineKeyType(id)

	var config models.QueryConfig

	if keyType == "id" {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			KeyName:   "id",
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	} else {
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_roles",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	}

	err := r.db.GetItem(ctx, config, &existingRole)
	if err != nil {
		r.logger.Errorf("Failed to get role by %s: %v", keyName, err)
		return fmt.Errorf("failed to get role by %s: %w", keyName, err)
	}

	if existingRole.ID == "" {
		return errors.New("role not found")
	}

	err = r.db.DeleteItem(ctx, r.config.DynamoDBTablePrefix+"_roles", "id", existingRole.ID)
	if err != nil {
		r.logger.Errorf("Failed to delete role: %v", err)
		return err
	}

	r.logger.Infof("Role deleted successfully: %s", existingRole.ID)
	return nil
}

func (r *RoleRepository) GetRolesByStatus(status models.RoleStatus) ([]*models.Role, error) {
	ctx := context.Background()
	var roles []*models.Role

	err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_roles", "status-index", "status", string(status), &roles)
	if err != nil {
		r.logger.Errorf("Failed to get roles by status: %v", err)
		return nil, err
	}

	return roles, nil
}