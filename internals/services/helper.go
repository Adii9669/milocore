package services

import (
	"chat-server/internals/db/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

// Helper Function
func (s *channelService) getMembership(
	ctx context.Context,
	channelID, userID uuid.UUID,
) (*models.ChannelMember, error) {

	member, err := s.channelRepo.FindMember(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}

	if member == nil {
		return nil, ErrMemberNotFound
	}

	return member, nil
}

func (s *channelService) requireRole(
	ctx context.Context,
	channelID, userID uuid.UUID,
	allowed ...models.ChannelRole,
) (*models.ChannelMember, error) {

	member, err := s.getMembership(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}

	for _, role := range allowed {
		if member.Role == role {
			return member, nil
		}
	}

	return nil, errors.New("Persmison Denied")
}
