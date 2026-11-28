package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type rejectFriendRepository struct {
	repository.FriendRepository
	relation    *models.Friend
	lookupLow   uuid.UUID
	lookupHigh  uuid.UUID
	deletedLow  uuid.UUID
	deletedHigh uuid.UUID
}

func (r *rejectFriendRepository) GetRelation(_ context.Context, low, high uuid.UUID) (*models.Friend, error) {
	r.lookupLow, r.lookupHigh = low, high
	return r.relation, nil
}

func (r *rejectFriendRepository) DeletePending(_ context.Context, low, high, sender uuid.UUID) (bool, error) {
	r.deletedLow, r.deletedHigh = low, high
	return r.relation.RequestedBy == sender, nil
}

func TestRejectIncomingFriendRequestUsesNormalizedPair(t *testing.T) {
	currentUserID := uuid.New()
	senderID := uuid.New()
	low, high := normalizePair(currentUserID, senderID)
	repo := &rejectFriendRepository{relation: &models.Friend{
		UserLowID: low, UserHighID: high, Status: StatusPending, RequestedBy: senderID,
	}}
	service := NewFriendService(nil, repo)

	if err := service.RejectRequest(context.Background(), currentUserID, senderID); err != nil {
		t.Fatalf("reject incoming request: %v", err)
	}
	if repo.lookupLow != low || repo.lookupHigh != high || repo.deletedLow != low || repo.deletedHigh != high {
		t.Fatal("request lookup and deletion must use the normalized user pair")
	}
}

func TestRejectFriendRequestDoesNotAllowSenderOrMissingRelation(t *testing.T) {
	currentUserID := uuid.New()
	otherUserID := uuid.New()
	low, high := normalizePair(currentUserID, otherUserID)
	repo := &rejectFriendRepository{relation: &models.Friend{
		UserLowID: low, UserHighID: high, Status: StatusPending, RequestedBy: currentUserID,
	}}
	service := NewFriendService(nil, repo)
	if err := service.RejectRequest(context.Background(), currentUserID, otherUserID); err == nil {
		t.Fatal("sender must not reject their own outgoing request")
	}

	repo.relation = nil
	if err := service.RejectRequest(context.Background(), currentUserID, otherUserID); !errors.Is(err, ErrRequestNotFound) {
		t.Fatalf("missing request should return ErrRequestNotFound, got %v", err)
	}
}

type cancelFriendRepository struct {
	repository.FriendRepository
	low, high, sender uuid.UUID
	removed           bool
}

func (r *cancelFriendRepository) DeletePending(_ context.Context, low, high, sender uuid.UUID) (bool, error) {
	r.low, r.high, r.sender = low, high, sender
	return r.removed, nil
}

func TestCancelFriendRequestUsesSenderAndPendingPair(t *testing.T) {
	sender, recipient := uuid.New(), uuid.New()
	low, high := normalizePair(sender, recipient)
	repo := &cancelFriendRepository{removed: true}
	service := NewFriendService(nil, repo)
	if err := service.CancelRequest(context.Background(), sender, recipient); err != nil {
		t.Fatal(err)
	}
	if repo.low != low || repo.high != high || repo.sender != sender {
		t.Fatal("cancellation must scope deletion to the normalized pair and original sender")
	}
	repo.removed = false
	if err := service.CancelRequest(context.Background(), recipient, sender); !errors.Is(err, ErrRequestNotFound) {
		t.Fatalf("missing or already accepted request: %v", err)
	}
}
