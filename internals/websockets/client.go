package websockets

import (
	"chat-server/internals/services"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second    // writeWait      → max time allowed for a write
	pongWait       = 60 * time.Second    // pongWait       → if no pong within 60s connection dies
	pingPeriod     = (pongWait * 9) / 10 // pingPeriod     → send ping every ~54s
	maxMessageSize = 8192                // maxMessageSize → max incoming message size (8KB)
)

type Client struct {
	conn *websocket.Conn
	send chan []byte

	userID uuid.UUID
	crews  map[uuid.UUID]bool

	hub *Hub

	ctx    context.Context
	cancel context.CancelFunc

	messageService services.MessageService
	limiter        *rate.Limiter
}

func NewClient(
	conn *websocket.Conn,
	userID uuid.UUID,
	hub *Hub,
	ctx context.Context,
	cancel context.CancelFunc,
	messageService services.MessageService,
) *Client {
	return &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		crews:          make(map[uuid.UUID]bool),
		hub:            hub,
		ctx:            ctx,
		cancel:         cancel,
		messageService: messageService,
		limiter:        rate.NewLimiter(10, 20),
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

	//this is to check the connection is still alive or not
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	//endless loop for reading message from the client
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("ws closed:", err)
			}
			break
		}

		if !c.limiter.Allow() {
			log.Printf("rate limit exceeded user=%s", c.userID)
			continue
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(data, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case "join_crew":
			c.hub.joinCrew(c, wsMsg.CrewID)
		case "leave_crew":
			c.hub.leaveCrew(c, wsMsg.CrewID)
		case "dm":
			c.handleDmMessage(wsMsg)
		case "crew_message":
			c.handleCrewMessage(wsMsg)
		case "typing":
			c.hub.broadcastTyping(c, wsMsg)
		case "delivered":
			c.handleDelivered(wsMsg)
		case "read":
			c.handleRead(wsMsg)
		default:
			log.Printf("unknow ws message type=%s from the user=%s", wsMsg.Type, c.userID)
		}

	}
}

// writePump
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		//this case if for singnalling when the client disconnects from the server
		//and using context we can signal our go routine to stop
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

			//sever sends the ticker and client will responsd to this ticker
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
