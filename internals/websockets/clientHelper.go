package websockets

import (
	"chat-server/internals/services"
	"context"
	"log"
	"time"
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

	savedMessage, err := c.messageService.HandleIncomingMessage(
		ctx,
		c.userID,
		incoming,
	)
	if err != nil {
		log.Println("failed to save message:", err)
		return
	}
	c.hub.sendAck(c.userID, savedMessage.Response.ID, "sent")
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
	c.hub.sendAck(c.userID, savedMessage.Response.ID, "sent")
	c.hub.route <- savedMessage
}
