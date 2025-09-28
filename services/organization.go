package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"regexp"
	"strings"
)

type OrganizationService struct {
	organizationRepo repository.OrganizationRepositoryInterface
	logger           logger.Logger
}

func NewOrganizationService(organizationRepo repository.OrganizationRepositoryInterface, logger logger.Logger) *OrganizationService {
	return &OrganizationService{
		organizationRepo: organizationRepo,
		logger:           logger,
	}
}

func (s *OrganizationService) CreateOrganization(ctx context.Context, organization *models.Organization, createdBy string) (*models.Organization, error) {
	if err := s.validateCreateOrganization(organization); err != nil {
		return nil, err
	}

	// Set system-generated fields
	organization.CreatedBy = createdBy
	organization.UpdatedBy = createdBy

	return s.organizationRepo.CreateOrganization(ctx, organization)
}

func (s *OrganizationService) validateCreateOrganization(organization *models.Organization) error {
	if organization == nil {
		return errors.New("organization is required")
	}

	if strings.TrimSpace(organization.Name) == "" {
		return errors.New("organization name is required")
	}

	if len(organization.Name) > 100 {
		return errors.New("organization name must be less than 100 characters")
	}

	if strings.TrimSpace(organization.Email) != "" {
		if !isValidEmail(organization.Email) {
			return errors.New("invalid email format")
		}
	}

	if strings.TrimSpace(organization.Phone) != "" {
		if !regexp.MustCompile(`^\+?[1-9]\d{1,14}$`).MatchString(organization.Phone) {
			return errors.New("invalid phone number format")
		}
	}

	if strings.TrimSpace(organization.Address) != "" {
		if len(organization.Address) > 200 {
			return errors.New("address must be less than 200 characters")
		}
	}

	if strings.TrimSpace(organization.City) != "" {
		if len(organization.City) < 2 || len(organization.City) > 50 {
			return errors.New("city must be between 2 and 50 characters")
		}
	}

	if strings.TrimSpace(organization.State) != "" {
		if len(organization.State) < 2 || len(organization.State) > 50 {
			return errors.New("state must be between 2 and 50 characters")
		}
	}

	if strings.TrimSpace(organization.Country) != "" {
		if len(organization.Country) < 2 || len(organization.Country) > 50 {
			return errors.New("country must be between 2 and 50 characters")
		}
	}

	if strings.TrimSpace(organization.PostalCode) != "" {
		if len(organization.PostalCode) < 3 || len(organization.PostalCode) > 20 {
			return errors.New("postal code must be between 3 and 20 characters")
		}
	}

	if strings.TrimSpace(organization.Industry) != "" {
		if len(organization.Industry) < 2 || len(organization.Industry) > 50 {
			return errors.New("industry must be between 2 and 50 characters")
		}
	}

	return nil
}

func isValidEmail(email string) bool {
	// Simple regex for email validation
	const emailRegex = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func (s *OrganizationService) GetOrganizations(key string) ([]*models.Organization, error) {
	return s.organizationRepo.GetOrganization(key)
}
