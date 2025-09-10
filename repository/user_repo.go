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

// GetUser retrieves a user by ID, email, or username
func (r *UserRepository) GetUser(key string) (*models.User, error) {
	user := models.User{}
	ctx := context.Background()
	
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
	return &user, nil
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
