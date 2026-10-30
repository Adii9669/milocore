package websockets

import (
	"encoding/json"
	"testing"

	"chat-server/internals/events"
	"github.com/google/uuid"
)

func TestPlanEventRoutesOnlyToExplicitAuthorizedRecipients(t *testing.T) {
	authorizedID, unauthorizedID := uuid.New(), uuid.New()
	planID, eventID := uuid.New(), uuid.New()
	authorized := &Client{userID: authorizedID, send: make(chan []byte, 1)}
	unauthorized := &Client{userID: unauthorizedID, send: make(chan []byte, 1)}
	hub := &Hub{users: map[uuid.UUID]map[*Client]bool{
		authorizedID:   {authorized: true},
		unauthorizedID: {unauthorized: true},
	}, unregister: make(chan *Client)}
	hub.handleRoutePlanEvent(events.PlanEvent{
		Event: events.PlanCreated, EventID: eventID, PlanID: planID,
		ChannelID: uuidPtr(uuid.New()), Recipients: []uuid.UUID{authorizedID, authorizedID},
	})
	select {
	case raw := <-authorized.send:
		var envelope struct {
			Event string        `json:"event"`
			Data  PlanEventData `json:"data"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			t.Fatalf("invalid event JSON: %v", err)
		}
		if envelope.Event != events.PlanCreated || envelope.Data.EventID != eventID.String() || envelope.Data.PlanID != planID.String() {
			t.Fatalf("unexpected plan event payload: %#v", envelope)
		}
	default:
		t.Fatal("authorized recipient did not receive plan event")
	}
	select {
	case raw := <-unauthorized.send:
		t.Fatalf("unauthorized user received plan event: %s", raw)
	default:
	}
}

func TestPlanEventIgnoresInvalidOrUnaddressedEvent(t *testing.T) {
	userID := uuid.New()
	client := &Client{userID: userID, send: make(chan []byte, 1)}
	hub := &Hub{users: map[uuid.UUID]map[*Client]bool{userID: {client: true}}}
	hub.handleRoutePlanEvent(events.PlanEvent{Event: events.PlanCreated, PlanID: uuid.New(), EventID: uuid.New()})
	hub.handleRoutePlanEvent(events.PlanEvent{Event: events.PlanCreated, PlanID: uuid.Nil, EventID: uuid.New(), Recipients: []uuid.UUID{userID}})
	if len(client.send) != 0 {
		t.Fatal("invalid or unaddressed event was delivered")
	}
}

func uuidPtr(id uuid.UUID) *uuid.UUID { return &id }
