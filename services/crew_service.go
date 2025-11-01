package services

import (
	"context"
	"errors"
	"fieldfuze-backend/models"
	"fieldfuze-backend/repository"
	"fieldfuze-backend/utils/logger"
	"strings"
)

type CrewService struct {
	crewRepo repository.CrewRepositoryInterface
	logger   logger.Logger
}

func NewCrewService(crewRepo repository.CrewRepositoryInterface, logger logger.Logger) *CrewService {
	return &CrewService{
		crewRepo: crewRepo,
		logger:   logger,
	}
}

func (s *CrewService) CreateCrew(ctx context.Context, req *models.CreateCrewRequest, createdBy string) (*models.Crew, error) {
	if err := s.validateCreateCrew(req); err != nil {
		return nil, err
	}

	crew := &models.Crew{
		Name:             req.Name,
		Description:      req.Description,
		LeadTechnicianId: req.LeadTechnicianId,
		MemberIds:        req.MemberIds,
		OrgID:            req.OrgID,
		Skills:           req.Skills,
		CreatedBy:        createdBy,
		IsActive:         true,
	}

	if crew.MemberIds == nil {
		crew.MemberIds = []string{}
	}
	if crew.Skills == nil {
		crew.Skills = []string{}
	}

	return s.crewRepo.CreateCrew(ctx, crew)
}

func (s *CrewService) validateCreateCrew(req *models.CreateCrewRequest) error {
	if req == nil {
		return errors.New("crew request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return errors.New("crew name is required")
	}

	if len(req.Name) < 2 || len(req.Name) > 100 {
		return errors.New("crew name must be between 2 and 100 characters")
	}

	if strings.TrimSpace(req.LeadTechnicianId) == "" {
		return errors.New("lead technician ID is required")
	}

	if strings.TrimSpace(req.OrgID) == "" {
		return errors.New("organization ID is required")
	}

	if len(req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	return nil
}

func (s *CrewService) GetCrews(filter *models.CrewFilter) ([]*models.Crew, error) {
	if filter == nil {
		filter = &models.CrewFilter{}
	}
	return s.crewRepo.GetCrewsByFilter(filter)
}

func (s *CrewService) GetCrewByID(id string) (*models.Crew, error) {
	crews, err := s.crewRepo.GetCrew(id)
	if err != nil {
		return nil, err
	}
	if len(crews) == 0 {
		return nil, errors.New("crew not found")
	}
	return crews[0], nil
}

func (s *CrewService) UpdateCrew(ctx context.Context, id string, req *models.UpdateCrewRequest) (*models.Crew, error) {
	if err := s.validateUpdateCrew(req); err != nil {
		return nil, err
	}

	existing, err := s.GetCrewByID(id)
	if err != nil {
		return nil, err
	}

	updatedCrew := *existing

	if req.Name != "" {
		updatedCrew.Name = req.Name
	}
	if req.Description != "" {
		updatedCrew.Description = req.Description
	}
	if req.LeadTechnicianId != "" {
		updatedCrew.LeadTechnicianId = req.LeadTechnicianId
	}
	if req.MemberIds != nil {
		updatedCrew.MemberIds = req.MemberIds
	}
	if req.Skills != nil {
		updatedCrew.Skills = req.Skills
	}
	if req.IsActive != nil {
		updatedCrew.IsActive = *req.IsActive
	}

	return s.crewRepo.UpdateCrew(id, &updatedCrew)
}

func (s *CrewService) validateUpdateCrew(req *models.UpdateCrewRequest) error {
	if req == nil {
		return errors.New("update request is required")
	}

	if req.Name != "" && (len(req.Name) < 2 || len(req.Name) > 100) {
		return errors.New("crew name must be between 2 and 100 characters")
	}

	if len(req.Description) > 500 {
		return errors.New("description must be less than 500 characters")
	}

	return nil
}

func (s *CrewService) DeleteCrew(id string) error {
	return s.crewRepo.DeleteCrew(id)
}

func (s *CrewService) GetCrewsByOrganization(orgID string, isActive *bool) ([]*models.Crew, error) {
	filter := &models.CrewFilter{
		OrgID: orgID,
	}
	if isActive != nil {
		filter.IsActive = isActive
	}
	return s.crewRepo.GetCrewsByFilter(filter)
}

func (s *CrewService) GetCrewsByLeadTechnician(leadTechnicianId string) ([]*models.Crew, error) {
	filter := &models.CrewFilter{
		LeadTechnicianId: leadTechnicianId,
	}
	return s.crewRepo.GetCrewsByFilter(filter)
}

func (s *CrewService) AddMemberToCrew(ctx context.Context, crewID, memberID string) (*models.Crew, error) {
	crew, err := s.GetCrewByID(crewID)
	if err != nil {
		return nil, err
	}

	for _, existingMember := range crew.MemberIds {
		if existingMember == memberID {
			return nil, errors.New("member already exists in crew")
		}
	}

	crew.MemberIds = append(crew.MemberIds, memberID)
	return s.crewRepo.UpdateCrew(crewID, crew)
}

func (s *CrewService) RemoveMemberFromCrew(ctx context.Context, crewID, memberID string) (*models.Crew, error) {
	crew, err := s.GetCrewByID(crewID)
	if err != nil {
		return nil, err
	}

	var updatedMembers []string
	memberFound := false
	for _, existingMember := range crew.MemberIds {
		if existingMember != memberID {
			updatedMembers = append(updatedMembers, existingMember)
		} else {
			memberFound = true
		}
	}

	if !memberFound {
		return nil, errors.New("member not found in crew")
	}

	crew.MemberIds = updatedMembers
	return s.crewRepo.UpdateCrew(crewID, crew)
}