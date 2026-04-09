package websockets

import (
	"chat-server/internals/events"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------route Message
func (h *Hub) handleRouteMessage(msg events.MessageEvent) {
	// log.Printf("Route check → ReceiverID = %v", msg.ReceiverID)
	if msg.ID == uuid.Nil {
		log.Println("empty message")
		return
	}

	payload, err := seralizeMessageEvent(msg)
	if err != nil {
		log.Print("ws marshal failed error:", err)
		return
	}
	//Debug
	// log.Println(string(payload))

	//here we handle the routing based on crewID/userID
	if msg.CrewID != nil {
		h.routeCrewMessage(payload, msg)
		return
	}
	if msg.ReceiverID != nil {
		h.routeDMMessage(payload, msg)
		return
	}
}

// -----------------------------------------------------------build message evenet
func seralizeMessageEvent(msg events.MessageEvent) ([]byte, error) {
	if msg.ID == uuid.Nil {
		return nil, fmt.Errorf("invalid message result")
	}
	//creating the Event of the message
	event := WSEvent{
		Event: EventMessage,
		Data: WSMessagePayload{
			ID:        msg.ID.String(),
			Type:      msg.Type,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
			Sender: SenderDTO{
				ID:   msg.SenderID.String(),
				Name: msg.SenderName,
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

// ----------------------------------------------------------- route dm messages
func (h *Hub) routeDMMessage(payload []byte, msg events.MessageEvent) {

	//checking the receiver exist or not in the connected client list (array)
	if msg.ReceiverID == nil {
		log.Printf("reciever id nil")
		return
	}
	receiverUUID := *msg.ReceiverID

	receiverClients, ok := h.users[receiverUUID]
	if !ok {
		return
	}

	log.Printf("Receiver has %d connections", len(receiverClients))

	//send message to the client (receiverClients)
	for client := range receiverClients {
		//using select to avoid the block's
		select {
		case client.send <- payload:
			//if client not active or stuck remove it from the map
		default:
			h.unregister <- client
		}
	}

}

// -------------------------------------------------------------- route crew messages
func (h *Hub) routeCrewMessage(payload []byte, msg events.MessageEvent) {

	if msg.CrewID == nil {
		return
	}
	// find all connections in that crew
	crewID := *msg.CrewID
	crewClients, ok := h.crews[crewID]
	if !ok {
		return
	}

	// send message to each connection
	for client := range crewClients {
		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}

}
