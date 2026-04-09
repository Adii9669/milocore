package services

import (
	"chat-server/internals/transport/dto"

	"github.com/google/uuid"
)

type MessageResult struct {
	Response   *dto.MessageResponse
	SenderID   string
	ReceiverID *uuid.UUID
	CrewID     *uuid.UUID
}
