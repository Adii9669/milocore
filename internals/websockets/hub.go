package websockets

import (
	"chat-server/internals/services"
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
	route     chan *services.MessageResult
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
		route:     make(chan *services.MessageResult),
	}

}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d  ",
				len(h.clients),
				len(h.users),
			)
			// for c := range h.clients {
			// 	log.Printf("   client connection → userID=%s", c.userID)
			// }

		case client := <-h.unregister:
			h.handleUnregister(client)

			// DEBUG
			log.Printf(
				"[WS] client registered | total connections=%d | users=%d",
				len(h.clients),
				len(h.users),
			)

		case msg := <-h.route:
			h.handleRouteMessage(msg)
		case data := <-h.broadcast:
			h.handleBroadcast(data)
			log.Printf("incoming WS payload: %s", string(data))
		}
	}
}
