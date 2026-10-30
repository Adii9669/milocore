package websockets

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

const (
	EventMessage = "message"
	EventAck     = "ack"
	EventTyping  = "typing"
)

func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	clients, ok := h.users[userID]
	return ok && len(clients) > 0
}

func (h *Hub) SendToUser(userID uuid.UUID, payload []byte) {
	if clients, ok := h.users[userID]; ok {
		for client := range clients {
			client.send <- payload
		}
	}
}

// ------------------------------------------------------------------ joinChannel
func (h *Hub) joinChannel(client *Client, channelID uuid.UUID) {

	if _, ok := h.channels[channelID]; !ok {
		h.channels[channelID] = make(map[*Client]bool)
	}

	h.channels[channelID][client] = true
	client.channels[channelID] = true
}

// ---------------------------------------------------------------leaveChannel
func (h *Hub) leaveChannel(client *Client, channelID uuid.UUID) {

	if channelClient, ok := h.channels[channelID]; ok {
		delete(channelClient, client)

		if len(channelClient) == 0 {
			delete(h.channels, channelID)
		}
	}
	delete(client.channels, channelID)
}

// ---------------------------------------------------------------typing
func (h *Hub) broadcastTyping(sender *Client, msg WSMessage) {

	//validate the channels exists
	channelClients, ok := h.channels[msg.ChannelID]
	if !ok {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		// log it, don't ignore
		return
	}

	for c := range channelClients {
		if c == sender {
			continue
		}
		//non blocking send
		select {
		case c.send <- payload:
		default:
			close(c.send)
			delete(h.clients, c)
		}
	}
}

func (h *Hub) sendMessageStatus(userID, msgID uuid.UUID, status string) {

	clients, ok := h.users[userID]
	if !ok {
		return
	}
	event := MessageStatusEvent{
		Type:      "message.status",
		MessageID: msgID.String(),
		Status:    status,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("marshal status failed", err)
		return
	}

	for client := range clients {

		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}
}
