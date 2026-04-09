package websockets

import (
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"chat-server/internals/workers"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
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
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "invalid userID", http.StatusBadRequest)
		return
	}

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

	deliveryWorker := &workers.DeliveryWorker{
		MessageService: messageservice,
		DeliveredChan:  hub.delivered,
	}

	deliveryWorker.Start(ctx)
	//run the pumps
	go client.writePump()
	go client.readPump()
}
