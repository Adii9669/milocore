package websockets

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (c *Client) handleReceipt(messageID uuid.UUID, status string) {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	sender, err := c.messageService.RecordDMReceipt(ctx, messageID, c.userID, status == "read")
	if err == nil && sender != uuid.Nil {
		c.hub.sendMessageStatus(sender, messageID, status)
	}
}
