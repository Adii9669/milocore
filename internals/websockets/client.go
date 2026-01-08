package websockets

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID string
	crews  map[string]bool
	hub    *Hub
}

func NewClient(conn *websocket.Conn, userID string, hub *Hub) *Client {
	return &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		crews:  make(map[string]bool),
		hub:    hub,
	}
}

// readPump
func (c *Client) readPump() {
	defer func() {
		// On exit, unregister client and close connection
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Println("ws closed:", err)
			}
			break
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.Type == "" || msg.ReceiverID == "" || msg.Content == "" {
			continue
		}

		//overwrite the senderID and update it to the userid from which the connection was made during the websocket connection
		//the make single source of truth the server don't have to trust on the payload userID send by the client
		//this is crusical for security
		msg.SenderID = c.userID

		c.hub.route <- OutgoingMessage{
			Type:       msg.Type,
			SenderID:   msg.SenderID,
			ReceiverID: msg.ReceiverID,
			Content:    msg.Content,
		}
	}
}

// writePump
func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
