package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/transport/dto"
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrRequestNotFound = errors.New("friend request not found")
	ErrAlreadyFriends  = errors.New("already friends")
	ErrInvalidRequest  = errors.New("invalid friend request state")
	ErrCannotAcceptOwn = errors.New("cannot accept your own request")
)

const (
	StatusPending  = "pending"
	StatusAccepted = "accepted"
)

func normalizePair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() < b.String() {
		return a, b
	}
	return b, a
}

type FriendService interface {
	SendFriendRequest(ctx context.Context, senderID uuid.UUID, targetID uuid.UUID) error
	AcceptFriendRequest(ctx context.Context, userID, otherID uuid.UUID) error
	GetFriends(ctx context.Context, userID uuid.UUID) ([]dto.FriendResponse, error)
}

type friendService struct {
	userRepo repository.UserRepository
	frndRepo repository.FriendRepository
}

func NewFriendService(
	userRepo repository.UserRepository,
	frndRepo repository.FriendRepository,
) FriendService {
	return &friendService{
		userRepo: userRepo,
		frndRepo: frndRepo,
	}
}

func (s *friendService) SendFriendRequest(
	ctx context.Context,
	senderID uuid.UUID,
	targetID uuid.UUID,
) error {

	//check if user and targetID not same
	if senderID == targetID {
		return errors.New("cannot send request to your self")
	}

	low, high := normalizePair(senderID, targetID)

	existing, err := s.frndRepo.FindBYPair(ctx, low, high)
	if err == nil && existing != nil {

		switch existing.Status {
		case StatusAccepted:
			return ErrAlreadyFriends
		case StatusPending:
			return ErrAlreadyFriends
		}
	}

	friend := &models.Friend{
		ID:          uuid.New(),
		UserLowID:   low,
		UserHighID:  high,
		Status:      StatusPending,
		RequestedBy: senderID,
	}

	return s.frndRepo.Create(ctx, friend)
}

func (s *friendService) AcceptFriendRequest(
	ctx context.Context,
	userID uuid.UUID,
	otherID uuid.UUID,
) error {

	low, high := normalizePair(userID, otherID)

	friend, err := s.frndRepo.FindBYPair(ctx, low, high)
	if err != nil {
		return ErrRequestNotFound
	}

	if friend.Status != StatusPending {
		return ErrRequestNotFound
	}

	if friend.RequestedBy == userID {
		return ErrCannotAcceptOwn
	}

	friend.Status = StatusAccepted

	return s.frndRepo.Update(ctx, friend)
}

func (s *friendService) GetFriends(
	ctx context.Context,
	userID uuid.UUID) ([]dto.FriendResponse, error) {

	rows, err := s.frndRepo.GetAcceptedByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))

	for _, f := range rows {
		if f.UserLowID == userID {
			ids = append(ids, f.UserHighID)
		} else {
			ids = append(ids, f.UserLowID)
		}
	}

	users, err := s.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FriendResponse, 0, len(users))

	for _, u := range users {
		result = append(result, dto.FriendResponse{
			ID:   u.ID.String(),
			Name: u.Name,
		})
	}

	return result, nil
}
