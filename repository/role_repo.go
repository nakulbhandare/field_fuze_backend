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

// RoleRepository implements RoleRepositoryInterface
type RoleRepository struct {
	db     dal.DatabaseClientInterface
	config *models.Config
	logger logger.Logger
}

func NewRoleRepository(db dal.DatabaseClientInterface, cfg *models.Config, log logger.Logger) *RoleRepository {
	return &RoleRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *RoleRepository) CreateRoleAssignment(ctx context.Context, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error) {
	r.logger.Infof("Creating role assignment: %s", roleAssignment.RoleName)

	existingRole := &models.RoleAssignment{}
	err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_role", "name-index", "role_name", roleAssignment.RoleName, &[]*models.RoleAssignment{existingRole})
	if err == nil && existingRole.RoleID != "" {
		return nil, errors.New("role with this name already exists")
	}

	now := time.Now()
	roleAssignment.RoleID = utils.GenerateUUID()
	roleAssignment.AssignedAt = now

	fmt.Println("roles ::::", dal.PrintPrettyJSON(roleAssignment))

	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_role", roleAssignment)
	if err != nil {
		r.logger.Errorf("Failed to create role assignment: %v", err)
		return nil, err
	}

	r.logger.Infof("Role assignment created successfully: %s", roleAssignment.RoleID)
	return roleAssignment, nil
}

func (r *RoleRepository) GetRoleAssignments(key string) ([]*models.RoleAssignment, error) {
	ctx := context.Background()

	if key == "" {
		var roleAssignments []*models.RoleAssignment
		tableName := r.config.DynamoDBTablePrefix + "_role"

		r.logger.Infof("Scanning %s table for all role assignments", tableName)

		err := r.db.ScanTable(ctx, tableName, &roleAssignments)
		if err != nil {
			r.logger.Errorf("Failed to scan role assignments table: %v", err)
			return nil, fmt.Errorf("failed to get all role assignments: %w", err)
		}

		r.logger.Infof("Found %d role assignments", len(roleAssignments))
		return roleAssignments, nil
	}

	roleAssignment := models.RoleAssignment{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_role",
		KeyName:   "role_id",
		KeyValue:  key,
		KeyType:   models.StringType,
	}

	r.logger.Infof("Querying %s table with role_id: %s", r.config.DynamoDBTablePrefix+"_role", key)

	err := r.db.GetItem(ctx, config, &roleAssignment)
	if err != nil {
		r.logger.Errorf("Failed to get role assignment by ID: %v", err)
		return nil, fmt.Errorf("failed to get role assignment by ID: %w", err)
	}

	return []*models.RoleAssignment{&roleAssignment}, nil
}

func (r *RoleRepository) GetRoleAssignmentsByStatus(status string) ([]*models.RoleAssignment, error) {
	ctx := context.Background()

	var roleAssignments []*models.RoleAssignment
	tableName := r.config.DynamoDBTablePrefix + "_role"

	r.logger.Infof("Scanning %s table for role assignments with status: %s", tableName, status)

	err := r.db.ScanTable(ctx, tableName, &roleAssignments)
	if err != nil {
		r.logger.Errorf("Failed to get role assignments by status: %v", err)
		return nil, fmt.Errorf("failed to get role assignments by status: %w", err)
	}

	// Since RoleAssignment doesn't have status field, return all for now
	// You may want to filter based on your business logic
	r.logger.Infof("Found %d role assignments for status filter", len(roleAssignments))
	return roleAssignments, nil
}

func (r *RoleRepository) UpdateRoleAssignment(id string, roleAssignment *models.RoleAssignment) (*models.RoleAssignment, error) {
	ctx := context.Background()
	r.logger.Infof("Updating role assignment: %s", id)

	// Check if role assignment exists
	existing, err := r.GetRoleAssignments(id)
	if err != nil || len(existing) == 0 {
		return nil, errors.New("role assignment not found")
	}

	// Update the role assignment
	roleAssignment.RoleID = id
	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_role", roleAssignment)
	if err != nil {
		r.logger.Errorf("Failed to update role assignment: %v", err)
		return nil, err
	}

	r.logger.Infof("Role assignment updated successfully: %s", id)
	return roleAssignment, nil
}

func (r *RoleRepository) DeleteRoleAssignment(id string) error {
	ctx := context.Background()
	r.logger.Infof("Deleting role assignment: %s", id)

	// Check if role assignment exists
	existing, err := r.GetRoleAssignments(id)
	if err != nil || len(existing) == 0 {
		return errors.New("role assignment not found")
	}

	tableName := r.config.DynamoDBTablePrefix + "_role"
	err = r.db.DeleteItem(ctx, tableName, "role_id", id)
	if err != nil {
		r.logger.Errorf("Failed to delete role assignment: %v", err)
		return fmt.Errorf("failed to delete role assignment: %w", err)
	}

	r.logger.Infof("Role assignment deleted successfully: %s", id)
	return nil
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
