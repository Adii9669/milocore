package mapper

import (
	"chat-server/internals/db/models"
	"chat-server/internals/transport/dto"
)

func ToDMMessageResponse(msg *models.Message, authUserID string) dto.MessageResponse {
	isMine := msg.SenderID == authUserID
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "dm",
		Content:   *msg.Content,
		IsMine:    &isMine,
		CreatedAt: msg.CreatedAt,
	}
}

func ToCrewMessageResponse(msg *models.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "crew",
		Content:   *msg.Content,
		CreatedAt: msg.CreatedAt,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID,
			Name: msg.SenderName,
		},
	}
}
