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

	// Save to database
	err = r.db.PutItem(ctx, r.config.DynamoDBTablePrefix+"_users", user)
	if err != nil {
		r.logger.Errorf("Failed to create user: %v", err)
		return nil, err
	}

	r.logger.Infof("User created successfully: %s", user.ID)
	return user, nil
}
