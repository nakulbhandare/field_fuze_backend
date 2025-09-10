package repository

import (
	"fieldfuze-backend/dal"
	"fieldfuze-backend/models"

	"github.com/bytedance/gopkg/util/logger"
)

type Repository struct {
	User *UserRepository
}

func NewRepository(db *dal.DynamoDBClient, cfg *models.Config, log logger.Logger) *Repository {
	return &Repository{
		User: NewUserRepository(db, cfg, log),
	}
}
