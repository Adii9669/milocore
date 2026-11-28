package websockets

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestInviteInvalidationIsTargetedAndDeduplicated(t *testing.T) {
	hub := NewHub()
	recipient := &Client{userID: uuid.New(), send: make(chan []byte, 2)}
	unrelated := &Client{userID: uuid.New(), send: make(chan []byte, 2)}
	hub.handleRegister(recipient)
	hub.handleRegister(unrelated)
	channelID := uuid.New()
	go hub.PublishGroupInvitesChanged(channelID, recipient.userID, recipient.userID)
	(<-hub.roomAction)()
	if len(recipient.send) != 1 || len(unrelated.send) != 0 {
		t.Fatal("unexpected invite event recipients")
	}
	var event struct {
		Event string `json:"event"`
		Data  struct {
			ChannelID string `json:"channelId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(<-recipient.send, &event); err != nil ||
		event.Event != "group_invites.changed" || event.Data.ChannelID != channelID.String() {
		t.Fatal("invalid invite invalidation")
	}
}
