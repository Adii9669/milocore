package websockets

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"time"

	"chat-server/internals/db"
	"chat-server/internals/db/models"
	"chat-server/internals/utils"
	"chat-server/middleware"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}

	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

type IncomingMessage struct {
	Content string `json:"content"`
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan Message
	UserID   string `json:"userId"`
	Username string `json:"username"`
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

	//get the details of the user who connected to it
	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	var err error

	identifier := claims.Username

	if emailRegex.MatchString(identifier) {
		log.Printf("DEBUG: Identifier '%s' is an email. Searching DB by email.", identifier)

		if result := db.DB.First(&user, "email = ?", identifier); result.Error != nil {
			err = result.Error
		}
	} else {
		log.Printf("DEBUG: Identifier '%s' is a username. Searching DB by username.", identifier)
		// Otherwise, assume it's a username
		if result := db.DB.First(&user, "username = ?", identifier); result.Error != nil {
			err = result.Error
		}
	}

	//check if the username happens to be nill or null we will not
	var username string
	if user.Name != nil {
		username = *user.Name
	}

	log.Printf("Connecting the %s...", claims.Username)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan Message),
		UserID:   user.ID,
		Username: username,
	}

	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer c.conn.Close()
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		//unmarshall the content came from the frontend
		var incomingMessage IncomingMessage
		if err := json.Unmarshal(rawMessage, &incomingMessage); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue // Skip malformed messages
		}

		// FIX 3: Create a structured Message and send it to the hub.
		message := Message{
			UserID:   c.UserID,
			Username: c.Username,
			Content:  incomingMessage.Content,
		}
		c.hub.broadcast <- message
	}
}
