package services

import (
	"chat-server/internals/db/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

// Helper Function
func (s *crewService) getMembership(
	ctx context.Context,
	crewID, userID uuid.UUID,
) (*models.CrewMember, error) {

	member, err := s.crewRepo.FindMember(ctx, crewID, userID)
	if err != nil {
		return nil, err
	}

	if member == nil {
		return nil, ErrMemberNotFound
	}

	return member, nil
}

func (s *crewService) requireRole(
	ctx context.Context,
	crewID, userID uuid.UUID,
	allowed ...models.CrewRole,
) (*models.CrewMember, error) {

	member, err := s.getMembership(ctx, crewID, userID)
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
