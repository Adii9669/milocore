package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidRole      = errors.New("invalid role")
	ErrNameEmpty        = errors.New("Name can not be empty")
	ErrOwnerNotExist    = errors.New("Owner does not exist")
	ErrAlreadyExist     = errors.New("already exist")
	ErrFailedToCreate   = errors.New("failed to create crew")
	ErrCrewNotFound     = errors.New("crew not found")
	ErrMemberNotFound   = errors.New("member not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrAlreadyMember    = errors.New("already a member")
)

type CrewService interface {
	//Crew Method
	CreateCrew(ctx context.Context, requesterID uuid.UUID, name string) (dto.CrewResponse, error)
	GetCrews(ctx context.Context, requesterID uuid.UUID) ([]dto.CrewResponse, error)
	DeleteCrew(ctx context.Context, crewID, ownerID uuid.UUID) error
	RenameCrew(ctx context.Context, crewID, requesterID uuid.UUID, newName string) error

	//Member
	AddMember(ctx context.Context, crewID, requesterID, targetID uuid.UUID) error
	GetMembers(ctx context.Context, crewID, ownerID uuid.UUID) (dto.CrewMembersResponse, error)
	UpdateMemberRole(ctx context.Context, crewID, requesterID, targetID uuid.UUID, newRole models.CrewRole) error
	RemoveMember(ctx context.Context, crewID, requesterID, targetID uuid.UUID) error
}

type crewService struct {
	db       *gorm.DB
	crewRepo repository.CrewRepository
	userRepo repository.UserRepository
}

func NewCrewService(
	db *gorm.DB,
	crewRepo repository.CrewRepository,
	userRepo repository.UserRepository) CrewService {
	return &crewService{
		db:       db,
		userRepo: userRepo,
		crewRepo: crewRepo,
	}
}

func (s *crewService) CreateCrew(ctx context.Context, requesterID uuid.UUID, name string) (dto.CrewResponse, error) {

	if strings.TrimSpace(name) == "" {
		return dto.CrewResponse{}, ErrNameEmpty
	}

	exist, err := s.userRepo.ExistByID(ctx, requesterID)
	if err != nil {
		return dto.CrewResponse{}, err
	}
	if !exist {
		return dto.CrewResponse{}, ErrOwnerNotExist
	}

	newCrew := &models.Crew{
		Name:    name,
		OwnerID: requesterID,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := s.crewRepo.CreateCrew(ctx, tx, newCrew); err != nil {
			return err
		}

		ownerMember := &models.CrewMember{
			CrewID: newCrew.ID,
			UserID: requesterID,
			Role:   models.RoleOwner,
		}

		if err := s.crewRepo.CreateMember(ctx, tx, ownerMember); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return dto.CrewResponse{}, ErrFailedToCreate
	}

	return dto.CrewResponse{
		ID:        newCrew.ID,
		Name:      newCrew.Name,
		OwnerID:   newCrew.OwnerID,
		CreatedAt: newCrew.CreatedAt,
	}, nil
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

func (s *crewService) DeleteCrew(
	ctx context.Context, crewID, requesterID uuid.UUID) error {

	//check User
	userExists, err := s.userRepo.ExistByID(ctx, requesterID)
	if err != nil {
		return err
	}
	if !userExists {
		return ErrUserNotFound
	}

	requester, err := s.getMembership(ctx, crewID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != models.RoleOwner {
		return ErrPermissionDenied
	}
	return s.crewRepo.DeleteCrew(ctx, crewID)
}

func (s *crewService) AddMember(
	ctx context.Context,
	crewID uuid.UUID,
	requesterID uuid.UUID,
	targetUserID uuid.UUID,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		//Check crew exists
		exists, err := s.crewRepo.ExistByID(ctx, crewID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrCrewNotFound
		}

		//Check requester membership
		requesterMember, err := s.crewRepo.FindMember(ctx, crewID, requesterID)
		if err != nil {
			return ErrMemberNotFound
		}
		if requesterMember == nil {
			return ErrMemberNotFound
		}
		// 3️⃣ Permission check
		if requesterMember.Role != models.RoleOwner &&
			requesterMember.Role != models.RoleAdmin {
			return ErrPermissionDenied
		}

		// 4️⃣ Check target user exists
		userExists, err := s.userRepo.ExistByID(ctx, targetUserID)
		if err != nil {
			return err
		}
		if !userExists {
			return ErrUserNotFound
		}

		// 5️⃣ Prevent duplicate membership
		// existing, err := s.crewRepo.FindMember(ctx, crewID, targetUserID)
		// if err != nil {
		// 	return err
		// }
		// if existing != nil {
		// 	return ErrAlreadyMember
		// }

		// 6️⃣ Insert new member
		newMember := &models.CrewMember{
			CrewID: crewID,
			UserID: targetUserID,
			Role:   models.RoleMember,
		}

		if err := s.crewRepo.CreateMember(ctx, tx, newMember); err != nil {
			//this logic is temporary need to be updated later
			if repository.IsUniqueViolation(err) {
				return ErrAlreadyMember
			}
			return err
		}

		return nil
	})
}

func (s *crewService) RemoveMember(ctx context.Context, crewID, requesterID, targetID uuid.UUID) error {

	// Get requester membership
	requester, err := s.getMembership(ctx, crewID, requesterID)
	if err != nil {
		return err
	}

	//Get target membership
	target, err := s.getMembership(ctx, crewID, targetID)
	if err != nil {
		return err
	}

	//Prevent self-removal
	if requesterID == targetID {
		return ErrPermissionDenied
	}

	// Nobody can remove owner
	if target.Role == models.RoleOwner {
		return ErrPermissionDenied
	}

	//Member cannot remove anyone
	if requester.Role == models.RoleMember {
		return ErrPermissionDenied
	}

	//Admin cannot remove admin
	if requester.Role == models.RoleAdmin &&
		target.Role == models.RoleAdmin {
		return ErrPermissionDenied
	}

	//Safe to remove
	return s.crewRepo.RemoveMember(ctx, crewID, targetID)
}

func (s *crewService) UpdateMemberRole(ctx context.Context, crewID,
	requesterID, targetID uuid.UUID, newRole models.CrewRole) error {

	// requester must be owner
	requester, err := s.getMembership(ctx, crewID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != models.RoleOwner {
		return ErrPermissionDenied
	}

	// target must exist
	target, err := s.getMembership(ctx, crewID, targetID)
	if err != nil {
		return err
	}

	if target.Role == newRole {
		return nil
	}

	// cannot update owner
	if target.Role == models.RoleOwner {
		return ErrPermissionDenied
	}

	// cannot assign owner here
	if newRole == models.RoleOwner {
		return ErrPermissionDenied
	}

	// validate role
	if !newRole.IsValid() {
		return ErrInvalidRole
	}

	return s.crewRepo.UpdateRole(ctx, crewID, targetID, newRole)

}

// RenameCrew
func (s *crewService) RenameCrew(ctx context.Context,
	crewID, requesterID uuid.UUID, newName string) error {

	// requester
	requester, err := s.getMembership(ctx, crewID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != models.RoleOwner &&
		requester.Role != models.RoleAdmin {
		return ErrPermissionDenied
	}

	if strings.TrimSpace(newName) == "" {
		return ErrNameEmpty
	}

	return s.crewRepo.RenameCrew(ctx, crewID, newName)
}

// get Members
func (s *crewService) GetMembers(ctx context.Context, crewID, ownerID uuid.UUID) (dto.CrewMembersResponse, error) {

	//requrest is the member and the Owner
	_, err := s.getMembership(ctx, crewID, ownerID)
	if err != nil {
		return dto.CrewMembersResponse{}, err
	}

	members, err := s.crewRepo.FindMembersByCrewID(ctx, crewID)
	if err != nil {
		return dto.CrewMembersResponse{}, err
	}

	result := make([]dto.MemberResponse, 0, len(members))

	for _, m := range members {
		result = append(result, dto.MemberResponse{
			UserID: m.UserID,
			Role:   string(m.Role),
		})
	}

	return dto.CrewMembersResponse{
		CrewID:      crewID,
		MemberCount: len(result),
		Members:     result,
	}, nil
}
