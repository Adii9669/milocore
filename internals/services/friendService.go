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
	SendFriendRequest(
		ctx context.Context,
		senderID uuid.UUID,
		targetID uuid.UUID,
	) error

	AcceptFriendRequest(
		ctx context.Context,
		userID,
		otherID uuid.UUID,
	) error

	GetFriends(
		ctx context.Context,
		userID uuid.UUID,
	) ([]dto.FriendResponse, error)

	RemoveFriend(
		ctx context.Context,
		userID uuid.UUID,
		friendID uuid.UUID,
	) error

	GetRequests(
		ctx context.Context,
		userID uuid.UUID,
		reqType string,
	) ([]dto.FriendResponse, error)

	RejectRequest(
		ctx context.Context,
		requesterID uuid.UUID,
		userID uuid.UUID,
	) error
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

func (s *friendService) RemoveFriend(
	ctx context.Context,
	userID uuid.UUID,
	frienID uuid.UUID,
) error {

	low, high := normalizePair(userID, frienID)

	friendship, err := s.frndRepo.GetRelation(ctx, low, high)
	if err != nil {
		return errors.New("friendship not found")
	}

	if friendship.Status != "accepted" {
		return errors.New("cannot remove non-accepted relation")
	}

	return s.frndRepo.Delete(ctx, low, high)
}

func (s *friendService) GetRequests(
	ctx context.Context,
	userID uuid.UUID,
	reqType string,
) ([]dto.FriendResponse, error) {
	switch reqType {

	case "outgoing":
		requests, err := s.frndRepo.GetOutgoingRequests(ctx, userID)
		if err != nil {
			return nil, err
		}
		result := make([]dto.FriendResponse, 0)
		for _, relation := range requests {
			otherID := relation.UserLowID
			if otherID == userID {
				otherID = relation.UserHighID
			}
			otherUser, err := s.userRepo.FindByID(ctx, otherID)
			if err != nil {
				return nil, err
			}
			result = append(result, dto.FriendResponse{
				ID:        otherUser.ID.String(),
				Name:      otherUser.Name,
				Status:    relation.Status,
				CreatedAt: relation.CreatedAt,
			})
		}
		return result, nil

	default:
		requests, err := s.frndRepo.GetIncomingRequests(ctx, userID)
		if err != nil {
			return nil, err
		}
		result := make([]dto.FriendResponse, 0)
		for _, relation := range requests {
			otherID := relation.UserLowID
			if otherID == userID {
				otherID = relation.UserHighID
			}
			otherUser, err := s.userRepo.FindByID(ctx, otherID)
			if err != nil {
				return nil, err
			}
			result = append(result, dto.FriendResponse{
				ID:        otherUser.ID.String(),
				Name:      otherUser.Name,
				Status:    relation.Status,
				CreatedAt: relation.CreatedAt,
			})
		}
		return result, nil
	}
}

func (s *friendService) RejectRequest(
	ctx context.Context,
	requesterID uuid.UUID,
	userID uuid.UUID,
) error {

	req, err := s.frndRepo.GetRelation(ctx, userID, requesterID)
	if err != nil {
		return err
	}

	if req.Status != "pending" {
		return errors.New("request not pending")
	}

	if req.RequestedBy == userID {
		return errors.New("sender cannot reject")
	}

	return s.frndRepo.Delete(ctx, userID, requesterID)
}
