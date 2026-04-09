package websockets

import (
	"context"
	"log"
	"time"
)

// -----------------------------------------------handle Delivered
func (c *Client) handleDelivered(msg WSMessage) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	err := c.messageService.MarkDelivered(ctx, *msg.MessageID)
	if err != nil {
		log.Println("failed to mark delivered:", err)
		return
	}

	// notify sender
	c.hub.sendMessageStatus(c.userID, *msg.MessageID, "delivered")
}

// --------------------------------------------------------handleRead
func (c *Client) handleRead(msg WSMessage) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	err := c.messageService.MarkRead(ctx, *msg.MessageID, c.userID)
	if err != nil {
		log.Println("failed to mark read:", err)
		return
	}

	c.hub.sendMessageStatus(c.userID, *msg.MessageID, "read")
}
