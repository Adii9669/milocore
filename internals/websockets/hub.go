package websockets

import (
	"chat-server/internals/events"
	"log"

	"github.com/google/uuid"
)

// here in struct means we are defining the type of the channels how it going to be there name and type
type Hub struct {
	clients    map[*Client]bool //every websockets connections
	register   chan *Client
	unregister chan *Client

	users map[uuid.UUID]map[*Client]bool //set of Clients connections
	crews map[uuid.UUID]map[*Client]bool

	broadcast chan []byte
	route     chan events.MessageEvent
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

		users: make(map[uuid.UUID]map[*Client]bool),
		crews: make(map[uuid.UUID]map[*Client]bool),

		broadcast: make(chan []byte, 256),
		route:     make(chan events.MessageEvent, 256),
		delivered: make(chan string, 256),
		read:      make(chan events.ReadEvent, 256),
	}

}

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
		case msg := <-h.route:
			h.handleRouteMessage(msg)

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
