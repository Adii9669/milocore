package websockets

import (
	"chat-server/internals/events"
	"testing"

	"github.com/google/uuid"
)

func TestDirectMessageRoutingEchoesPersistedMessageToSenderAndReceiver(t *testing.T) {
	senderID, receiverID := uuid.New(), uuid.New()
	sender := &Client{userID: senderID, send: make(chan []byte, 1)}
	receiver := &Client{userID: receiverID, send: make(chan []byte, 1)}
	hub := &Hub{users: map[uuid.UUID]map[*Client]bool{
		senderID:   {sender: true},
		receiverID: {receiver: true},
	}, unregister: make(chan *Client)}
	payload := []byte(`{"event":"message","data":{"id":"persisted-id"}}`)
	hub.routeDMMessage(payload, events.MessageEvent{SenderID: senderID, ReceiverID: &receiverID})

	for name, client := range map[string]*Client{"sender": sender, "receiver": receiver} {
		select {
		case delivered := <-client.send:
			if string(delivered) != string(payload) {
				t.Errorf("%s received unexpected payload: %s", name, delivered)
			}
		default:
			t.Errorf("%s did not receive persisted message event", name)
		}
	}
}
