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
	ErrFailedToCreate   = errors.New("failed to create channel")
	ErrChannelNotFound  = errors.New("channel not found")
	ErrMemberNotFound   = errors.New("member not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrAlreadyMember    = errors.New("already a member")
)

type ChannelService interface {
	//Channel Method
	CreateChannel(ctx context.Context, requesterID uuid.UUID, name string) (dto.ChannelResponse, error)
	GetChannels(ctx context.Context, requesterID uuid.UUID) ([]dto.ChannelResponse, error)
	DeleteChannel(ctx context.Context, channelID, ownerID uuid.UUID) error
	RenameChannel(ctx context.Context, channelID, requesterID uuid.UUID, newName string) error

	//Member
	AddMember(ctx context.Context, channelID, requesterID, targetID uuid.UUID) error
	GetMembers(ctx context.Context, channelID, ownerID uuid.UUID) (dto.ChannelMembersResponse, error)
	UpdateMemberRole(ctx context.Context, channelID, requesterID, targetID uuid.UUID, newRole models.ChannelRole) error
	RemoveMember(ctx context.Context, channelID, requesterID, targetID uuid.UUID) error
}

type channelService struct {
	db          *gorm.DB
	channelRepo repository.ChannelRepository
	userRepo    repository.UserRepository
}

func NewChannelService(
	db *gorm.DB,
	channelRepo repository.ChannelRepository,
	userRepo repository.UserRepository) ChannelService {
	return &channelService{
		db:          db,
		userRepo:    userRepo,
		channelRepo: channelRepo,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, requesterID uuid.UUID, name string) (dto.ChannelResponse, error) {

	if strings.TrimSpace(name) == "" {
		return dto.ChannelResponse{}, ErrNameEmpty
	}

	exist, err := s.userRepo.ExistByID(ctx, requesterID)
	if err != nil {
		return dto.ChannelResponse{}, err
	}
	if !exist {
		return dto.ChannelResponse{}, ErrOwnerNotExist
	}

	newChannel := &models.Channel{
		Name:    name,
		OwnerID: requesterID,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := s.channelRepo.CreateChannel(ctx, tx, newChannel); err != nil {
			return err
		}

		ownerMember := &models.ChannelMember{
			ChannelID: newChannel.ID,
			UserID:    requesterID,
			Role:      models.RoleOwner,
		}

		if err := s.channelRepo.CreateMember(ctx, tx, ownerMember); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return dto.ChannelResponse{}, ErrFailedToCreate
	}

	return dto.ChannelResponse{
		ID:        newChannel.ID,
		Name:      newChannel.Name,
		OwnerID:   newChannel.OwnerID,
		Role:      string(models.RoleOwner),
		CreatedAt: newChannel.CreatedAt,
	}, nil
}

func (s *channelService) GetChannels(ctx context.Context, ownerID uuid.UUID) ([]dto.ChannelResponse, error) {

	channels, err := s.channelRepo.FindForUser(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ChannelResponse, 0, len(channels))

	for _, c := range channels {
		membership, err := s.channelRepo.FindMember(ctx, c.ID, ownerID)
		if err != nil {
			return nil, err
		}
		if membership == nil {
			continue
		}
		result = append(result, dto.ChannelResponse{
			ID:      c.ID,
			Name:    c.Name,
			OwnerID: c.OwnerID,
			Role:    string(membership.Role),
		})
	}

	return result, nil
}

func (s *channelService) DeleteChannel(
	ctx context.Context, channelID, requesterID uuid.UUID) error {

	//check User
	userExists, err := s.userRepo.ExistByID(ctx, requesterID)
	if err != nil {
		return err
	}
	if !userExists {
		return ErrUserNotFound
	}

	requester, err := s.getMembership(ctx, channelID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != models.RoleOwner {
		return ErrPermissionDenied
	}
	return s.channelRepo.DeleteChannel(ctx, channelID)
}

func (s *channelService) AddMember(
	ctx context.Context,
	channelID uuid.UUID,
	requesterID uuid.UUID,
	targetUserID uuid.UUID,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		//Check channel exists
		exists, err := s.channelRepo.ExistByID(ctx, channelID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrChannelNotFound
		}

		//Check requester membership
		requesterMember, err := s.channelRepo.FindMember(ctx, channelID, requesterID)
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
		// existing, err := s.channelRepo.FindMember(ctx, channelID, targetUserID)
		// if err != nil {
		// 	return err
		// }
		// if existing != nil {
		// 	return ErrAlreadyMember
		// }

		// 6️⃣ Insert new member
		newMember := &models.ChannelMember{
			ChannelID: channelID,
			UserID:    targetUserID,
			Role:      models.RoleMember,
		}

		if err := s.channelRepo.CreateMember(ctx, tx, newMember); err != nil {
			//this logic is temporary need to be updated later
			if repository.IsUniqueViolation(err) {
				return ErrAlreadyMember
			}
			return err
		}

		return nil
	})
}

func (s *channelService) RemoveMember(ctx context.Context, channelID, requesterID, targetID uuid.UUID) error {

	// Get requester membership
	requester, err := s.getMembership(ctx, channelID, requesterID)
	if err != nil {
		return err
	}

	//Get target membership
	target, err := s.getMembership(ctx, channelID, targetID)
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
	return s.channelRepo.RemoveMember(ctx, channelID, targetID)
}

func (s *channelService) UpdateMemberRole(ctx context.Context, channelID,
	requesterID, targetID uuid.UUID, newRole models.ChannelRole) error {

	// requester must be owner
	requester, err := s.getMembership(ctx, channelID, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != models.RoleOwner {
		return ErrPermissionDenied
	}

	// target must exist
	target, err := s.getMembership(ctx, channelID, targetID)
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

	return s.channelRepo.UpdateRole(ctx, channelID, targetID, newRole)

}

// RenameChannel
func (s *channelService) RenameChannel(ctx context.Context,
	channelID, requesterID uuid.UUID, newName string) error {

	// requester
	requester, err := s.getMembership(ctx, channelID, requesterID)
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

	return s.channelRepo.RenameChannel(ctx, channelID, newName)
}

// get Members
func (s *channelService) GetMembers(ctx context.Context, channelID, ownerID uuid.UUID) (dto.ChannelMembersResponse, error) {

	//requrest is the member and the Owner
	_, err := s.getMembership(ctx, channelID, ownerID)
	if err != nil {
		return dto.ChannelMembersResponse{}, err
	}

	members, err := s.channelRepo.FindMembersByChannelID(ctx, channelID)
	if err != nil {
		return dto.ChannelMembersResponse{}, err
	}

	result := make([]dto.MemberResponse, 0, len(members))
	userIDs := make([]uuid.UUID, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return dto.ChannelMembersResponse{}, err
	}
	usernames := make(map[uuid.UUID]string, len(users))
	for _, user := range users {
		usernames[user.ID] = user.Name
	}

	for _, m := range members {
		result = append(result, dto.MemberResponse{
			UserID:   m.UserID,
			Username: usernames[m.UserID],
			Role:     string(m.Role),
		})
	}

	return dto.ChannelMembersResponse{
		ChannelID:   channelID,
		MemberCount: len(result),
		Members:     result,
	}, nil
}
