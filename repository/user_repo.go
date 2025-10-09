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

type UserRepository struct {
	db     *dal.DynamoDBClient
	config *models.Config
	logger logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *dal.DynamoDBClient, cfg *models.Config, log logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		config: cfg,
		logger: log,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	fmt.Println("Creating user:", utils.PrintPrettyJSON(user))
	existingUser := &models.User{}
	err := r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_users", "email-index", "email", user.Email, &[]*models.User{existingUser})
	if err == nil && existingUser.ID != "" {
		return nil, errors.New("user with this email already exists")
	}

	// Check if username already exists
	existingUser = &models.User{}
	err = r.db.QueryByIndex(ctx, r.config.DynamoDBTablePrefix+"_users", "username-index", "username", user.Username, &[]*models.User{existingUser})
	if err == nil && existingUser.ID != "" {
		return nil, errors.New("user with this username already exists")
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = utils.GenerateUUID()
	user.Status = "active"
	user.Roles = []models.RoleAssignment{} // Initialize empty roles array
	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		r.logger.Errorf("Failed to hash password: %v", err)
		return nil, err
	}
	user.Password = hashedPassword

	// Save to database
	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_users", user)
	if err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return nil, err
	}

	r.logger.Infof("User created successfully: %s", user.ID)
	return user, nil
}

// GetUser retrieves users by ID, email, username, or returns all users if key is empty
func (r *UserRepository) GetUser(key string) ([]*models.User, error) {
	ctx := context.Background()

	// If key is empty, return all users
	if key == "" {
		var users []*models.User
		tableName := r.config.DynamoDBTablePrefix + "_users"

		fmt.Printf("Scanning %s table for all users\n", tableName)

		err := r.db.ScanTable(ctx, tableName, &users)
		if err != nil {
			r.logger.Errorf("Failed to scan users table: %v", err)
			return nil, fmt.Errorf("failed to get all users: %w", err)
		}

		fmt.Printf("Found %d users\n", len(users))
		return users, nil
	}

	// Single user lookup
	user := models.User{}

	// Determine the key type and set up query config accordingly
	keyType, indexName, keyName := r.determineKeyType(key)

	var config models.QueryConfig

	if keyType == "id" {
		// For ID, use direct GetItem without index
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_users",
			KeyName:   "id",
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	} else {
		// For email/username, use secondary index
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_users",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  key,
			KeyType:   models.StringType,
		}
	}

	fmt.Printf("Querying %s table with %s: %s\n", r.config.DynamoDBTablePrefix, keyName, key)

	err := r.db.GetItem(ctx, config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get user by %s: %w", keyName, err)
	}

	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	fmt.Println("User found:", utils.PrintPrettyJSON(user))
	return []*models.User{&user}, nil
}

// determineKeyType determines if the key is an ID, email, or username
func (r *UserRepository) determineKeyType(key string) (keyType, indexName, keyName string) {
	// UUID pattern for ID (format: 8-4-4-4-12 characters)
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	isUUID, _ := regexp.MatchString(uuidPattern, strings.ToLower(key))

	if isUUID {
		return "id", "", "id"
	} else if strings.Contains(key, "@") {
		// Contains @ symbol, likely an email
		return "email", "email-index", "email"
	} else {
		// Assume it's a username
		return "username", "username-index", "username"
	}
}

func (r *UserRepository) UpdateUser(id string, user *models.User) (*models.User, error) {
	ctx := context.Background()

	// Fetch existing user using the same logic as GetUser
	existingUser := models.User{}

	// Determine the key type and set up query config accordingly
	keyType, indexName, keyName := r.determineKeyType(id)

	var config models.QueryConfig

	if keyType == "id" {
		// For ID, use direct GetItem without index
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_users",
			KeyName:   "id",
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	} else {
		// For email/username, use secondary index
		config = models.QueryConfig{
			TableName: r.config.DynamoDBTablePrefix + "_users",
			IndexName: indexName,
			KeyName:   keyName,
			KeyValue:  id,
			KeyType:   models.StringType,
		}
	}

	err := r.db.GetItem(ctx, config, &existingUser)
	if err != nil {
		r.logger.Errorf("Failed to get user by %s: %v", keyName, err)
		return nil, fmt.Errorf("failed to get user by %s: %w", keyName, err)
	}

	if existingUser.ID == "" {
		return nil, errors.New("user not found")
	}

	// Prepare update fields
	updates := make(map[string]interface{})

	if user.FirstName != "" {
		updates["first_name"] = user.FirstName
	}
	if user.LastName != "" {
		updates["last_name"] = user.LastName
	}
	if user.Phone != nil {
		updates["phone"] = user.Phone
	}
	if user.Status != "" {
		updates["status"] = user.Status
	}
	// Role and Roles fields are not allowed to be updated via UpdateUser
	// Use dedicated role assignment methods instead
	if user.Password != "" {
		// Hash new password
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			r.logger.Errorf("Failed to hash password: %v", err)
			return nil, err
		}
		updates["password_hash"] = hashedPassword
	}
	updates["updated_at"] = time.Now()

	// Save updates
	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", existingUser.ID, updates)
	if err != nil {
		r.logger.Errorf("Failed to update user: %v", err)
		return nil, err
	}

	// Update the existing user object for return
	if user.FirstName != "" {
		existingUser.FirstName = user.FirstName
	}
	if user.LastName != "" {
		existingUser.LastName = user.LastName
	}
	if user.Phone != nil {
		existingUser.Phone = user.Phone
	}
	if user.Status != "" {
		existingUser.Status = user.Status
	}
	// Role and Roles fields are not updated via UpdateUser method
	if user.Password != "" {
		existingUser.Password = updates["password_hash"].(string)
	}
	existingUser.UpdatedAt = time.Now()

	r.logger.Infof("User updated successfully: %s", existingUser.ID)
	return &existingUser, nil
}

// AssignRoles assigns roles to a user
func (r *UserRepository) AssignRoles(ctx context.Context, userID string, roleAssignments []models.RoleAssignment) (*models.User, error) {
	// Get existing user
	user := models.User{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_users",
		KeyName:   "id",
		KeyValue:  userID,
		KeyType:   models.StringType,
	}

	err := r.db.GetItem(ctx, config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	// Set assigned timestamp for new roles
	now := time.Now()
	for i := range roleAssignments {
		roleAssignments[i].AssignedAt = now
	}

	// Update roles
	updates := map[string]interface{}{
		"roles":      roleAssignments,
		"updated_at": now,
	}

	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", userID, updates)
	if err != nil {
		r.logger.Errorf("Failed to assign roles to user: %v", err)
		return nil, err
	}

	// Update user object for return
	user.Roles = roleAssignments
	user.UpdatedAt = now

	r.logger.Infof("Roles assigned successfully to user: %s", userID)
	return &user, nil
}

// AddRoleToUser adds a single role to a user's existing roles
func (r *UserRepository) AddRoleToUser(ctx context.Context, userID string, roleAssignment models.RoleAssignment) (*models.User, error) {
	// Get existing user
	user := models.User{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_users",
		KeyName:   "id",
		KeyValue:  userID,
		KeyType:   models.StringType,
	}

	err := r.db.GetItem(ctx, config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	// Initialize roles if nil
	if user.Roles == nil {
		user.Roles = []models.RoleAssignment{}
	}

	// Check if role already exists
	for i, existingRole := range user.Roles {
		if existingRole.RoleID == roleAssignment.RoleID {
			// Update existing role
			roleAssignment.AssignedAt = time.Now()
			user.Roles[i] = roleAssignment

			updates := map[string]interface{}{
				"roles":      user.Roles,
				"updated_at": time.Now(),
			}

			err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", userID, updates)
			if err != nil {
				r.logger.Errorf("Failed to update role for user: %v", err)
				return nil, err
			}

			user.UpdatedAt = time.Now()
			r.logger.Infof("Role updated successfully for user: %s", userID)
			return &user, nil
		}
	}

	// Add new role
	roleAssignment.AssignedAt = time.Now()
	user.Roles = append(user.Roles, roleAssignment)

	updates := map[string]interface{}{
		"roles":      user.Roles,
		"updated_at": time.Now(),
	}

	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", userID, updates)
	if err != nil {
		r.logger.Errorf("Failed to add role to user: %v", err)
		return nil, err
	}

	user.UpdatedAt = time.Now()
	r.logger.Infof("Role added successfully to user: %s", userID)
	return &user, nil
}

// AssignRoleToUser assigns an existing role by ID to a user
func (r *UserRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) (*models.User, error) {
	// Get existing user
	user := models.User{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_users",
		KeyName:   "id",
		KeyValue:  userID,
		KeyType:   models.StringType,
	}

	err := r.db.GetItem(ctx, config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	// Get role details from role repository
	role := models.RoleAssignment{}
	roleConfig := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_role",
		KeyName:   "role_id",
		KeyValue:  roleID,
		KeyType:   models.StringType,
	}

	err = r.db.GetItem(ctx, roleConfig, &role)
	if err != nil {
		r.logger.Errorf("Failed to get role by ID: %v", err)
		return nil, errors.New("role not found")
	}

	if role.RoleID == "" {
		return nil, errors.New("role not found")
	}

	// Initialize roles if nil
	if user.Roles == nil {
		user.Roles = []models.RoleAssignment{}
	}

	// Check if user already has this role
	for _, existingRole := range user.Roles {
		if existingRole.RoleID == roleID {
			return nil, errors.New("user already has this role")
		}
	}

	// Add role to user with assigned timestamp
	role.AssignedAt = time.Now()
	user.Roles = append(user.Roles, role)

	updates := map[string]interface{}{
		"roles":      user.Roles,
		"updated_at": time.Now(),
	}

	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", userID, updates)
	if err != nil {
		r.logger.Errorf("Failed to assign role to user: %v", err)
		return nil, err
	}

	user.UpdatedAt = time.Now()
	r.logger.Infof("Role %s assigned successfully to user: %s", roleID, userID)
	return &user, nil
}

// RemoveRoleFromUser removes a role from a user
func (r *UserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) (*models.User, error) {
	// Get existing user
	user := models.User{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_users",
		KeyName:   "id",
		KeyValue:  userID,
		KeyType:   models.StringType,
	}

	err := r.db.GetItem(ctx, config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if user.ID == "" {
		return nil, errors.New("user not found")
	}

	// Find and remove the role
	updatedRoles := []models.RoleAssignment{}
	roleFound := false

	for _, role := range user.Roles {
		if role.RoleID != roleID {
			updatedRoles = append(updatedRoles, role)
		} else {
			roleFound = true
		}
	}

	if !roleFound {
		return nil, errors.New("role not found for user")
	}

	user.Roles = updatedRoles

	updates := map[string]interface{}{
		"roles":      user.Roles,
		"updated_at": time.Now(),
	}

	err = r.db.UpdateItem(ctx, r.config.DynamoDBTablePrefix+"_users", "id", userID, updates)
	if err != nil {
		r.logger.Errorf("Failed to remove role from user: %v", err)
		return nil, err
	}

	user.UpdatedAt = time.Now()
	r.logger.Infof("Role removed successfully from user: %s", userID)
	return &user, nil
}
