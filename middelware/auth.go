package middelware

import (
	"fieldfuze-backend/models"
	"fieldfuze-backend/utils"
	"fieldfuze-backend/utils/logger"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)



// JWTManager handles JWT token operations
type JWTManager struct {
	Config *models.Config
	Logger logger.Logger
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *models.Config, log logger.Logger) *JWTManager {
	return &JWTManager{
		Config: cfg,
		Logger: log,
	}
}

// GenerateToken generates a JWT token for a user
func (j *JWTManager) GenerateToken(user *models.User) (string, error) {
	// Create claims
	claims := models.JWTClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		Status:   user.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // JTI (JWT ID)
			Subject:   user.ID,
			Issuer:    j.Config.AppName,
			Audience:  jwt.ClaimStrings{j.Config.AppName},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Config.JWTExpiresIn)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(j.Config.JWTSecret))
	if err != nil {
		j.Logger.Errorf("Failed to sign JWT token: %v", err)
		return "", err
	}

	j.Logger.Debugf("Generated JWT token for user: %s", user.ID)

	fmt.Println(tokenString, utils.PrintPrettyJSON(tokenString))
	return tokenString, nil
}
