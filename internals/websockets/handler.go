package websockets

import (
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(
	hub *Hub,
	messageservice services.MessageService,
	w http.ResponseWriter,
	r *http.Request,

) {

	//get user Id
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusBadRequest)
		return
	}

	claims, err := utils.ValidateToken(cookie.Value)
	if err != nil {
		http.Error(w, "userId is required", http.StatusBadRequest)
	}

	userID := claims.UserID

	//upgrade the connections
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade failed:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	//create the client
	client := NewClient(conn, userID, hub, ctx, cancel, messageservice)

	//register the client in the hub
	hub.register <- client

	//run the pumps
	go client.writePump()
	go client.readPump()
}
