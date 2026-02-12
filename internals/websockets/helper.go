package websockets

import (
	"chat-server/internals/services"
	"encoding/json"
	"log"
)

// helper function
// register client
func (h *Hub) handleRegister(client *Client) {
	// Track global clients
	h.clients[client] = true

	// Track per-user connections
	if _, ok := h.users[client.userID]; !ok {
		h.users[client.userID] = make(map[*Client]bool)
	}
	h.users[client.userID][client] = true
}

// unregister client
func (h *Hub) handleUnregister(client *Client) {
	// Remove from global clients
	if _, ok := h.clients[client]; ok {
		// removing client from hub ke clients se
		delete(h.clients, client)
		close(client.send)
	}

	// Remove from user map
	if userClients, ok := h.users[client.userID]; ok {
		delete(userClients, client)
		if len(userClients) == 0 {
			delete(h.users, client.userID)
		}
	}

	// Remove from all crews
	for crewID := range client.crews {
		if crewClients, ok := h.crews[crewID]; ok {
			delete(crewClients, client)
			if len(crewClients) == 0 {
				delete(h.crews, crewID)
			}
		}
	}
}

// braodcast messages
func (h *Hub) handleBroadcast(data []byte) {
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

// dm messages
func (h *Hub) routeDMMessage(msg *services.MessageResult) {
	receiverClients, ok := h.users[*msg.ReceiverID]
	if !ok {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal DM", err)
		return
	}

	for client := range receiverClients {
		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}

	log.Printf("Sending ack")
	// notify sender that message was delivered
	h.sendAck(msg.SenderID, msg.Response.ID, "delivered")
}

// crew messages
func (h *Hub) routeCrewMessage(msg *services.MessageResult) {
	crewClients, ok := h.crews[*msg.CrewID]
	if !ok {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Println("Failed to marshal Crew", err)
		return
	}

	for client := range crewClients {
		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}

	// notify sender that message was delivered
	h.sendAck(msg.SenderID, msg.Response.ID, "delivered")

}

func (h *Hub) routeMessage(msg *services.MessageResult) {
	if msg.CrewID != nil {
		h.routeCrewMessage(msg)
		return
	}
	if msg.ReceiverID != nil {
		h.routeDMMessage(msg)
		return
	}

}

// acknowledge messages
func (h *Hub) sendAck(senderID, msgID, status string) {
	clients, ok := h.users[senderID]
	if !ok {
		return
	}

	ack := AckMessage{
		Type:      "ack",
		MessageID: msgID,
		Status:    status,
	}

	payload, err := json.Marshal(ack)
	if err != nil {
		log.Println("marshal ack failed:", err)
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
