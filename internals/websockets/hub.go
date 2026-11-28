package websockets

import (
	"chat-server/internals/events"
	"encoding/json"
	"log"

	"github.com/google/uuid"
)

// here in struct means we are defining the type of the channels how it going to be there name and type
type pushPublisher interface {
	Message(events.MessageEvent)
	Plan(events.PlanEvent)
}

type Hub struct {
	push       pushPublisher
	clients    map[*Client]bool //every websockets connections
	register   chan *Client
	unregister chan *Client
	roomAction chan func()

	users    map[uuid.UUID]map[*Client]bool //set of Clients connections
	channels map[uuid.UUID]map[*Client]bool

	broadcast chan []byte
	route     chan events.MessageEvent
	deletion  chan events.MessageDeletionEvent
	planEvent chan events.PlanEvent
	// route     chan *services.MessageResult

	delivered chan string
	read      chan events.ReadEvent
}

// here we are making those channels in the constructor
// defining channels
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		roomAction: make(chan func()),

		users:    make(map[uuid.UUID]map[*Client]bool),
		channels: make(map[uuid.UUID]map[*Client]bool),

		broadcast: make(chan []byte, 256),
		route:     make(chan events.MessageEvent, 256),
		deletion:  make(chan events.MessageDeletionEvent, 256),
		planEvent: make(chan events.PlanEvent, 256),
		delivered: make(chan string, 256),
		read:      make(chan events.ReadEvent, 256),
	}

}

// PublishMessage routes a message created outside the websocket handler, such as
// an uploaded attachment, through the same live delivery path.
func (h *Hub) PublishMessage(message events.MessageEvent)               { h.route <- message }
func (h *Hub) PublishMessageDeletion(event events.MessageDeletionEvent) { h.deletion <- event }
func (h *Hub) PublishPlanEvent(event events.PlanEvent) {
	if h == nil || h.planEvent == nil {
		return
	}
	h.planEvent <- event
}

// SetPushPublisher is called during startup, before Run begins.
func (h *Hub) SetPushPublisher(publisher pushPublisher) { h.push = publisher }

func (h *Hub) Run() {
	for {
		select {
		//case when register
		case client := <-h.register:
			h.handleRegister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d  ",
				len(h.clients),
				len(h.users),
			)

			//case when unregister a client
		case client := <-h.unregister:
			h.handleUnregister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d",
				len(h.clients),
				len(h.users),
			)

			//case to handle the mssages
		case action := <-h.roomAction:
			action()
		case msg := <-h.route:
			h.handleRouteMessage(msg)
			if h.push != nil {
				h.push.Message(msg)
			}
		case event := <-h.deletion:
			h.handleRouteMessageDeletion(event)
		case event := <-h.planEvent:
			h.handleRoutePlanEvent(event)
			if h.push != nil {
				h.push.Plan(event)
			}

			//case to handle the broadcast
		case data := <-h.broadcast:
			h.handleBroadcast(data)
			log.Printf("incoming WS payload: %s", string(data))
		}
	}
}

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

	// Remove from all channels
	for channelID := range client.channels {
		if channelClients, ok := h.channels[channelID]; ok {
			delete(channelClients, client)
			if len(channelClients) == 0 {
				delete(h.channels, channelID)
			}
		}
	}

}

// RevokeChannel removes all sockets for a removed member. A nil user revokes the entire channel.
func (h *Hub) RevokeChannel(channelID, userID uuid.UUID) {
	if h == nil {
		return
	}
	h.roomAction <- func() {
		for client := range h.channels[channelID] {
			if userID == uuid.Nil || client.userID == userID {
				h.leaveChannel(client, channelID)
				payload, _ := json.Marshal(WSEvent{Event: "channel.join_denied", Data: map[string]string{"channelId": channelID.String()}})
				h.sendToClient(client, payload)
			}
		}
	}
}

// PublishGroupInvitesChanged sends an invalidation to the invite parties and channel managers.
func (h *Hub) PublishGroupInvitesChanged(channel uuid.UUID, users ...uuid.UUID) {
	if h == nil {
		return
	}
	payload, _ := json.Marshal(WSEvent{Event: "group_invites.changed", Data: map[string]string{"channelId": channel.String()}})
	h.roomAction <- func() {
		seen := make(map[uuid.UUID]bool)
		for _, user := range users {
			if seen[user] {
				continue
			}
			seen[user] = true
			for client := range h.users[user] {
				h.sendToClient(client, payload)
			}
		}
	}
}
