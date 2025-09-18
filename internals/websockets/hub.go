package websockets

import (
	"log"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	UserID   string `json:"id"`
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {

	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {

	for {
		select {
		//register the client on the websocket
		case client := <-h.register:
			h.clients[client] = true
			log.Println("A new client is connected . Total clients are = ", len(h.clients))

			//for unregistering the client from the clients channel
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("A client has disconnected ", len(h.clients))
			}

			//for messages broadcasting
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
