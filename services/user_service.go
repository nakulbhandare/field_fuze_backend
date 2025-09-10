package services

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
)

type UserService struct {
	ctx  context.Context
	repo *repository.UserRepository
}

func NewUserService(ctx context.Context, repo *repository.UserRepository) *UserService {
	return &UserService{
		ctx:  ctx,
		repo: repo,
	}
}

func (s *UserService) CreateUser(user *models.User) (*models.User, error) {
	return s.repo.CreateUser(s.ctx, user)
}
