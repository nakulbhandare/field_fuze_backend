package repository

import (
	"context"
	"errors"
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils"
	"fmt"
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

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := models.User{}
	config := models.QueryConfig{
		TableName: r.config.DynamoDBTablePrefix + "_users",
		IndexName: "email-index",
		KeyName:   "email",
		KeyValue:  email,
		KeyType:   models.StringType,
	}
	fmt.Println(r.config.DynamoDBTablePrefix, "email :: ", email)
	err := r.db.GetItem(context.Background(), config, &user)
	if err != nil {
		r.logger.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}
	fmt.Println("Users found:", utils.PrintPrettyJSON(user))
	if user.ID == "" {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
