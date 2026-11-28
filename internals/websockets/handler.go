package websockets

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/workers"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		switch r.Header.Get("Origin") {
		case "", "http://localhost:3000", "https://milonext.vercel.app", "https://milonext.onrender.com":
			return true
		default:
			return false
		}
	},
}

type ChannelMembership interface {
	IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error)
}

func ServeWS(
	hub *Hub,
	messageservice services.MessageService,
	membership ChannelMembership,
	w http.ResponseWriter,
	r *http.Request,
) {

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
	client.membership = membership

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
