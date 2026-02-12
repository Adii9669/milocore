package websockets

import (
	"chat-server/internals/services"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	userID         string
	crews          map[string]bool
	hub            *Hub
	ctx            context.Context
	cancel         context.CancelFunc
	messageService services.MessageService
}

func NewClient(
	conn *websocket.Conn,
	userID string,
	hub *Hub,
	ctx context.Context,
	cancel context.CancelFunc,
	messageService services.MessageService,
) *Client {
	return &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		crews:          make(map[string]bool),
		hub:            hub,
		ctx:            ctx,
		cancel:         cancel,
		messageService: messageService,
	}
}

// readPump
func (c *Client) readPump() {
	defer func() {
		// On exit, unregister client and close connection
		c.cancel()
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

		var incomingmsg services.IncomingMessage
		if err := json.Unmarshal(data, &incomingmsg); err != nil {
			continue
		}

		//context for per meesage
		msgCtx, cancel := context.WithTimeout(c.ctx, 5*time.Second)

		savedMsg, err := c.messageService.HandleIncomingMessage(
			msgCtx,
			c.userID,
			incomingmsg,
		)
		cancel()
		if err != nil {
			log.Println("failed to save message:", err)
			continue
		}

		//overwrite the senderID and update it to the userid from which the connection was made during the websocket connection
		//the make single source of truth the server don't have to trust on the payload userID send by the client
		//this is crusical for security

		c.hub.route <- savedMsg

	}
}

// writePump
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}
}

// func (c *Client) writePump() {
// 	defer c.conn.Close()
// 	for msg := range c.send {
// 		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
// 		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
// 			return
// 		}
// 	}
// }
