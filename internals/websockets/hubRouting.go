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

	//here we handle the routing based on channelID/userID
	if msg.ChannelID != nil {
		h.routeChannelMessage(payload, msg)
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
			ChannelID:  msg.ChannelID,
			FileURL:    msg.FileURL, FilePreviewURL: msg.FilePreviewURL, FileName: msg.FileName, FileSize: msg.FileSize, MimeType: msg.MimeType,
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
	if msg.ReceiverID == nil {
		log.Printf("receiver id nil")
		return
	}

	// Send the persisted event to both participants. The sender needs the
	// database ID to replace its optimistic client-side message ID.
	recipients := map[uuid.UUID]struct{}{
		msg.SenderID:    {},
		*msg.ReceiverID: {},
	}
	for userID := range recipients {
		for client := range h.users[userID] {
			select {
			case client.send <- payload:
			default:
				h.unregister <- client
			}
		}
	}
}

// -------------------------------------------------------------- route channel messages
func (h *Hub) routeChannelMessage(payload []byte, msg events.MessageEvent) {

	if msg.ChannelID == nil {
		return
	}
	// find all connections in that channel
	channelID := *msg.ChannelID
	channelClients, ok := h.channels[channelID]
	if !ok {
		return
	}

	// send message to each connection
	for client := range channelClients {
		select {
		case client.send <- payload:
		default:
			h.unregister <- client
		}
	}

}

func (h *Hub) handleRouteMessageDeletion(event events.MessageDeletionEvent) {
	payload, err := json.Marshal(WSEvent{Event: string(MessageDeleted), Data: WSMessageDeletionPayload{
		MessageID: event.MessageID.String(), SenderID: event.SenderID.String(), ReceiverID: event.ReceiverID, ChannelID: event.ChannelID,
	}})
	if err != nil {
		log.Print("ws marshal failed error:", err)
		return
	}
	if event.ChannelID != nil {
		for client := range h.channels[*event.ChannelID] {
			select {
			case client.send <- payload:
			default:
				h.unregister <- client
			}
		}
		return
	}
	recipients := map[uuid.UUID]bool{event.SenderID: true}
	if event.ReceiverID != nil {
		recipients[*event.ReceiverID] = true
	}
	for userID := range recipients {
		for client := range h.users[userID] {
			select {
			case client.send <- payload:
			default:
				h.unregister <- client
			}
		}
	}
}

func (h *Hub) handleRoutePlanEvent(event events.PlanEvent) {
	if event.Event == "" || event.EventID == uuid.Nil || event.PlanID == uuid.Nil || len(event.Recipients) == 0 {
		return
	}
	payload, err := json.Marshal(WSEvent{Event: event.Event, Data: PlanEventData{
		EventID: event.EventID.String(), PlanID: event.PlanID.String(), ChannelID: event.ChannelID,
		DMUserLowID: event.DMUserLowID, DMUserHighID: event.DMUserHighID,
	}})
	if err != nil {
		log.Print("ws plan event marshal failed: ", err)
		return
	}
	seen := make(map[uuid.UUID]struct{}, len(event.Recipients))
	for _, userID := range event.Recipients {
		if userID == uuid.Nil {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		for client := range h.users[userID] {
			select {
			case client.send <- payload:
			default:
				h.unregister <- client
			}
		}
	}
}
