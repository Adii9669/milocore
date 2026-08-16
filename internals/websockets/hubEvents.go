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

// ------------------------------------------------------------------ joinCrew
func (h *Hub) joinCrew(client *Client, crewID uuid.UUID) {

	if _, ok := h.crews[crewID]; !ok {
		h.crews[crewID] = make(map[*Client]bool)
	}

	h.crews[crewID][client] = true
	client.crews[crewID] = true
}

// ---------------------------------------------------------------leaveCrew
func (h *Hub) leaveCrew(client *Client, crewID uuid.UUID) {

	if crewClient, ok := h.crews[crewID]; ok {
		delete(crewClient, client)

		if len(crewClient) == 0 {
			delete(h.crews, crewID)
		}
	}
	delete(client.crews, crewID)
}

// ---------------------------------------------------------------typing
func (h *Hub) broadcastTyping(sender *Client, msg WSMessage) {

	//validate the crews exists
	crewClients, ok := h.crews[msg.CrewID]
	if !ok {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		// log it, don't ignore
		return
	}

	for c := range crewClients {
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
