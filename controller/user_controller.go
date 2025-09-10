package controller

import (
	"context"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fmt"
	"net/http"

	"fieldfuze-backend/utils/logger"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	ctx      context.Context
	userRepo *repository.UserRepository
	// jwtManager *auth.JWTManager
	logger logger.Logger
}

func NewUserController(ctx context.Context, userRepo *repository.UserRepository, logger logger.Logger) *UserController {
	return &UserController{
		ctx:      ctx,
		userRepo: userRepo,
		logger:   logger,
	}
}

// Register handles POST /api/v1/auth/register
// @Summary Register a new user
// @Description Create a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.User true "Registration request"
// @Success 201 {object} models.APIResponse "User registered successfully"
// @Failure 400 {object} models.APIResponse "Bad Request - Invalid registration data"
// @Failure 409 {object} models.APIResponse "Conflict - User already exists"
// @Failure 500 {object} models.APIResponse "Internal Server Error - Registration failed"
// @Router /auth/register [post]
func (h *UserController) Register(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON:", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Code:    http.StatusBadRequest,
			Message: "Invalid request",
			Error: &models.APIError{
				Type:    "ValidationError",
				Details: err.Error(),
			},
		})
		return
	}

	user, err := h.userRepo.CreateUser(h.ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create user", fmt.Errorf("error: %v", err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user",
			Error: &models.APIError{
				Type:    "DatabaseError",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Status:  "success",
		Code:    http.StatusCreated,
		Message: "User registered successfully",
		Data:    user,
	})
}
