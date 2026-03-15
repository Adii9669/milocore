package websockets

import (
	"chat-server/internals/services"
	"encoding/json"
	"log"
)

const (
	EventMessage = "message"
	EventAck     = "ack"
	EventTyping  = "typing"
)

// ---------------------------------------------------- register client
func (h *Hub) handleRegister(client *Client) {
	// Track global clients
	h.clients[client] = true

	// Track per-user connections
	if _, ok := h.users[client.userID]; !ok {
		h.users[client.userID] = make(map[*Client]bool)
	}
	h.users[client.userID][client] = true
}

// ------------------------------------------------- unregister client
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

// -----------------------------------------------------------build message evenet
func buildMessageEvent(msg *services.MessageResult) ([]byte, error) {
	event := WSEvent{
		Event: EventMessage,
		Data: WSMessagePayload{
			ID:        msg.Response.ID,
			Type:      msg.Response.Type,
			Content:   msg.Response.Content,
			CreatedAt: msg.Response.CreatedAt,
			Sender: SenderDTO{
				ID:   msg.Response.Sender.ID,
				Name: msg.Response.Sender.Name,
			},
			ReceiverID: msg.ReceiverID,
			CrewID:     msg.CrewID,
		},
	}

	return json.Marshal(event)
}

// --------------------------------------------------------- braodcast messages
func (h *Hub) handleBroadcast(data []byte) {
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

// ---------------------------------------------------------------route Message
func (h *Hub) handleRouteMessage(msg *services.MessageResult) {
	// log.Printf("Route check → ReceiverID = %v", msg.ReceiverID)
	payload, err := buildMessageEvent(msg)
	log.Println(string(payload))
	if err != nil {
		log.Print("ws marshal failed error:", err)
		return
	}

	if msg.CrewID != nil {
		h.routeCrewMessage(payload, msg)
		return
	}
	if msg.ReceiverID != nil {
		h.routeDMMessage(payload, msg)

		return
	}
}

// ----------------------------------------------------------- route dm messages
func (h *Hub) routeDMMessage(payload []byte, msg *services.MessageResult) {

	//checking the receiver exist or not in the connected client list (array)
	receiverClients, ok := h.users[*msg.ReceiverID]
	if !ok {
		return
	}

	log.Printf("Receiver has %d connections", len(receiverClients))

	//send message to the client (receiverClients)
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

// -------------------------------------------------------------- route crew messages
func (h *Hub) routeCrewMessage(payload []byte, msg *services.MessageResult) {

	// find all connections in that crew
	crewClients, ok := h.crews[*msg.CrewID]
	if !ok {
		return
	}

	// // convert message to JSON
	// payload, err := json.Marshal(msg)
	// if err != nil {
	// 	log.Println("Failed to marshal Crew", err)
	// 	return
	// }

	// send message to each connection
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

// ------------------------------------------------------------------------acknowledge messages
func (h *Hub) sendAck(senderID, msgID, status string) {
	clients, ok := h.users[senderID]
	if !ok {
		return
	}

	ack := AckMessage{
		Type:      EventAck,
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

// ------------------------------------------------------------------ joinCrew
func (h *Hub) joinCrew(client *Client, crewID string) {

	if _, ok := h.crews[crewID]; !ok {
		h.crews[crewID] = make(map[*Client]bool)
	}

	h.crews[crewID][client] = true
	client.crews[crewID] = true
}

// ---------------------------------------------------------------leaveCrew
func (h *Hub) leaveCrew(client *Client, crewID string) {

	if crewClient, ok := h.crews[crewID]; ok {
		delete(crewClient, client)

		if len(crewClient) == 0 {
			delete(h.crews, crewID)
		}
	}
	delete(client.crews, crewID)
}

// ---------------------------------------------------------------typing
func (h *Hub) broadcastTyping(client *Client, msg WSMessage) {

	crewClients, ok := h.crews[msg.CrewID]
	if !ok {
		return
	}

	payload, _ := json.Marshal(msg)

	for c := range crewClients {
		if c != client {
			c.send <- payload
		}
	}
}
