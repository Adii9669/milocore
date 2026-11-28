package invites

import (
	"errors"
	"testing"
	"time"

	"chat-server/internals/db/models"

	"github.com/google/uuid"
)

func TestOnlyInviteeCanAcceptOrDecline(t *testing.T) {
	invite := models.ChannelInvite{InviteeID: uuid.New(), InviterID: uuid.New()}
	for _, state := range []string{"accepted", "declined"} {
		if err := inviteResponsePermission(invite, invite.InviteeID, state); err != nil {
			t.Fatal(err)
		}
		if !errors.Is(inviteResponsePermission(invite, invite.InviterID, state), missing) {
			t.Fatalf("inviter was allowed to respond with %s", state)
		}
	}
}

func TestPendingInviteExpiresAtDeadlineAndFinalStatesCannotChange(t *testing.T) {
	deadline := time.Now()
	invite := models.ChannelInvite{Status: "pending", ExpiresAt: deadline}
	if err := pendingAt(invite, deadline.Add(-time.Nanosecond)); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(pendingAt(invite, deadline), expired) {
		t.Fatal("invite remained actionable at expiry")
	}
	for _, state := range []string{"accepted", "declined", "cancelled", "expired"} {
		invite.Status = state
		if !errors.Is(pendingAt(invite, deadline.Add(-time.Second)), closed) {
			t.Fatalf("final state %s could be changed", state)
		}
	}
}

func TestInviteLimitsArePerAccountAndRecover(t *testing.T) {
	limits := NewLimits()
	first, second := uuid.New(), uuid.New()
	now := time.Now()
	for i := 0; i < 6; i++ {
		if !limits.allow(first, now) {
			t.Fatal("burst was limited early")
		}
	}
	if limits.allow(first, now) || !limits.allow(second, now) || !limits.allow(first, now.Add(2*time.Second)) {
		t.Fatal("invite limit did not scope or replenish the burst")
	}
}
