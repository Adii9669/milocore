package websockets

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type subscriptionMembership bool

func (m subscriptionMembership) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return bool(m), nil
}

func TestChannelSubscriptionsRequireMembershipAndSupportRevocation(t *testing.T) {
	hub := NewHub()
	channelID := uuid.New()
	client := &Client{
		userID: uuid.New(), ctx: context.Background(), membership: subscriptionMembership(false),
		hub: hub, channels: make(map[uuid.UUID]bool), send: make(chan []byte, 1),
	}
	hub.handleRegister(client)

	apply := func(request func(), expected string) {
		t.Helper()
		go request()
		(<-hub.roomAction)()
		select {
		case raw := <-client.send:
			var response WSEvent
			if err := json.Unmarshal(raw, &response); err != nil || response.Event != expected {
				t.Fatalf("unexpected subscription result: %s", raw)
			}
		default:
			t.Fatal("missing subscription result")
		}
	}
	join := func() { client.changeSubscription(channelID, true) }
	apply(join, "channel.join_denied")
	if hub.channels[channelID][client] {
		t.Fatal("a non-member joined the channel")
	}

	client.membership = subscriptionMembership(true)
	apply(join, "channel.joined")
	if !hub.channels[channelID][client] {
		t.Fatal("an authorized member did not join")
	}

	apply(func() { hub.RevokeChannel(channelID, client.userID) }, "channel.join_denied")
	if hub.channels[channelID][client] || client.channels[channelID] {
		t.Fatal("revoked channel subscription survived")
	}
}

func TestWebSocketOriginCheck(t *testing.T) {
	for origin, allowed := range map[string]bool{
		"": true, "http://localhost:3000": true,
		"https://milonext.vercel.app": true, "https://unrelated.example": false,
	} {
		req := httptest.NewRequest("GET", "/api/ws", nil)
		req.Header.Set("Origin", origin)
		if upgrader.CheckOrigin(req) != allowed {
			t.Errorf("unexpected origin decision for %q", origin)
		}
	}
}
