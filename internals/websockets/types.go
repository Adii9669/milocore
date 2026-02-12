package websockets

import "chat-server/internals/transport/dto"

type RouteMessage struct {
	ID         string
	MesageType string
	SenderID   string
	ReceiverID *string
	CrewID     *string
	Payload    *dto.MessageResponse
}
