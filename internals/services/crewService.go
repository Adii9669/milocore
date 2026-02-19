package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

type CrewService interface {
	CreateCrew(ctx context.Context, ownerID uuid.UUID, name string) (dto.CrewResponse, error)
	GetCrews(ctx context.Context, ownerID uuid.UUID) ([]dto.CrewResponse, error)
	DeleteCrew(ctx context.Context, crewID, ownerID uuid.UUID) error
}

type crewService struct {
	crewRepo repository.CrewRepository
	userRepo repository.UserRepository
}

func NewCrewService(
	crewRepo repository.CrewRepository,
	userRepo repository.UserRepository) CrewService {
	return &crewService{
		userRepo: userRepo,
		crewRepo: crewRepo,
	}
}

var (
	ErrNameEmpty     = errors.New("Name can not be empty")
	ErrOwnerNotExist = errors.New("Owner does not exist")
)

func (s *crewService) CreateCrew(ctx context.Context, ownerID uuid.UUID, name string) (dto.CrewResponse, error) {

	if strings.TrimSpace(name) == "" {
		return dto.CrewResponse{}, ErrNameEmpty
	}

	//check is user exist
	exist, err := s.userRepo.ExistByID(ctx, ownerID)
	if err != nil {
		return dto.CrewResponse{}, err
	}
	if !exist {
		return dto.CrewResponse{}, ErrOwnerNotExist
	}

	//CreateCrew new crew modal
	newCrew := &models.Crew{
		Name:    name,
		OwnerID: ownerID,
	}

	if err := s.crewRepo.CreateCrew(ctx, newCrew); err != nil {
		return dto.CrewResponse{}, errors.New("Failed to create crew")
	}

	// auto add owner as member
	// if err := s.crewRepo.AddMember(ctx, crew.ID, ownerID); err != nil {
	// 	return dto.CrewResponse{}, err
	// }

	response := dto.CrewResponse{
		ID:        newCrew.ID,
		Name:      newCrew.Name,
		OwnerID:   newCrew.OwnerID,
		CreatedAt: newCrew.CreatedAt,
	}

	return response, nil
}

func (s *crewService) GetCrews(ctx context.Context, ownerID uuid.UUID) ([]dto.CrewResponse, error) {

	crews, err := s.crewRepo.FindForUser(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CrewResponse, 0, len(crews))

	for _, c := range crews {
		result = append(result, dto.CrewResponse{
			ID:      c.ID,
			Name:    c.Name,
			OwnerID: c.OwnerID,
		})
	}

	return result, nil
}

func (s *crewService) DeleteCrew(ctx context.Context, crewID, ownerID uuid.UUID) error {
	crew, err := s.crewRepo.FindByID(ctx, crewID)
	if err != nil {
		return errors.New("crew not found")
	}

	if crew.OwnerID != ownerID {
		return errors.New("only owner can delete crew")
	}
	return s.crewRepo.DeleteCrewByID(ctx, ownerID, crewID)
}
