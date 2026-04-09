package websockets

import (
	"chat-server/internals/services"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

// ------------------------------------------------------------------handle Crew Message
func (c *Client) handleCrewMessage(msg WSMessage) {

	incoming := services.IncomingMessage{
		Type:    "crew",
		CrewID:  &msg.CrewID,
		Content: msg.Content,
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	//saving the mssage Service is called here
	savedMessage, err := c.messageService.HandleIncomingMessage(
		ctx,
		c.userID,
		incoming,
	)
	//Check is no error is returned from service
	if err != nil {
		log.Printf("DEBUG DM result: %+v\n", savedMessage)
		log.Println("failed to save message:", err)
		return
	}

	//CHeck is Message is Empty or not
	if savedMessage.ID == uuid.Nil {
		log.Println("❌ savedMessage is nil before routing")
		return
	}
	c.hub.sendMessageStatus(c.userID, savedMessage.ID, "sent")
	c.hub.route <- savedMessage
}

// -------------------------------------------------------------------handle Dm Messgae
func (c *Client) handleDmMessage(msg WSMessage) {

	incoming := services.IncomingMessage{
		Type:       "dm",
		ReceiverID: &msg.ReceiverID,
		Content:    msg.Content,
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	savedMessage, err := c.messageService.HandleIncomingMessage(
		ctx,
		c.userID,
		incoming,
	)
	if err != nil {
		log.Println("failed to save message:", err)
		return
	}

	//CHeck is Message is Empty or not
	if savedMessage.ID == uuid.Nil {
		log.Printf("DEBUG DM result: %+v\n", savedMessage)
		log.Println("❌ savedMessage is nil before routing")
		return
	}
	c.hub.sendMessageStatus(c.userID, savedMessage.ID, "sent")
	c.hub.route <- savedMessage
}
