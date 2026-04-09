package mapper

import (
	"chat-server/internals/db/models"
	"chat-server/internals/transport/dto"

	"github.com/google/uuid"
)

func ToDMMessageResponse(msg *models.Message, currentUser uuid.UUID) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "dm",
		Content:   *msg.Content,
		IsMine:    &isMine,
		CreatedAt: msg.CreatedAt,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID.String(),
			Name: msg.Sender.Name,
		},
	}
}

func ToCrewMessageResponse(msg *models.Message, currentUser uuid.UUID) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "crew",
		Content:   *msg.Content,
		CreatedAt: msg.CreatedAt,
		IsMine:    &isMine,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID.String(),
			Name: msg.Sender.Name,
		},
	}
}
