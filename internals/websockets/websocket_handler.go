package websockets

import (
	"encoding/json"
	"log"
	"net/http"

	"chat-server/internals/utils"
	"chat-server/middleware"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan Message
	UserID string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this to your frontend origin
	},
}

// serveWs handles a WebSocket request from a client.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {

	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Connecting the %s...", claims.Name)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan Message),
		UserID: claims.Subject,
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		// FIX 2: Marshal the Message struct into JSON ([]byte) before sending.
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		json.NewEncoder(w).Encode(message)

		if err := w.Close(); err != nil {
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		// FIX 3: Create a structured Message and send it to the hub.
		message := Message{
			Content: string(rawMessage),
			UserID:  c.UserID,
		}
		c.hub.broadcast <- message
	}
}
