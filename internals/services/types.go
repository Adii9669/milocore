package services

import "chat-server/internals/transport/dto"

type MessageResult struct {
	Response   *dto.MessageResponse
	SenderID   string
	ReceiverID *string
	CrewID     *string
}
