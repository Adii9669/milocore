package websockets

import (
	"encoding/json"
	"log"
)

// here in struct means we are defining the type of the channels how it going to be there name and type
type Hub struct {
	clients    map[*Client]bool //every websockets connections
	register   chan *Client
	unregister chan *Client

	users map[string]map[*Client]bool //set of Clients connections
	crews map[string]map[*Client]bool

	broadcast chan []byte
	route     chan OutgoingMessage
}

// here we are making those channels in the constructor
// defining channels
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),

		users: make(map[string]map[*Client]bool),
		crews: make(map[string]map[*Client]bool),

		broadcast: make(chan []byte, 256),
		route:     make(chan OutgoingMessage),
	}

}

// helper function
func (h *Hub) handleRegister(client *Client) {
	// Track global clients
	h.clients[client] = true

	// Track per-user connections
	if _, ok := h.users[client.userID]; !ok {
		h.users[client.userID] = make(map[*Client]bool)
	}
	h.users[client.userID][client] = true
}

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

func (h *Hub) sendDirectMessage(msg OutgoingMessage) {
	clients, ok := h.users[msg.ReceiverID]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// client is slow or dead
			h.unregister <- client
		}
	}
}

func (h *Hub) sendCrewMessage(msg OutgoingMessage) {
	clients, ok := h.crews[msg.ReceiverID]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

func (h *Hub) handleBroadcast(data []byte) {
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			h.unregister <- client
		}
	}
}

func (h *Hub) handleMessage(msg OutgoingMessage) {
	// We will route messages here later
	switch msg.Type {
	case "dm":
		h.sendDirectMessage(msg)

	case "crew":
		h.sendCrewMessage(msg)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d",
				len(h.clients),
				len(h.users),
			)

		case client := <-h.unregister:
			h.handleUnregister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d",
				len(h.clients),
				len(h.users),
			)

		case msg := <-h.route:
			h.handleMessage(msg)
		case data := <-h.broadcast:
			h.handleBroadcast(data)
		}
	}
}
