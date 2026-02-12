package mapper

import (
	"chat-server/internals/db/models"
	"chat-server/internals/transport/dto"
)

func ToDMMessageResponse(msg *models.Message, currentUser string) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "dm",
		Content:   *msg.Content,
		IsMine:    &isMine,
		CreatedAt: msg.CreatedAt,
	}
}

func ToCrewMessageResponse(msg *models.Message, currentUser string) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "crew",
		Content:   *msg.Content,
		CreatedAt: msg.CreatedAt,
		IsMine:    &isMine,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID,
			Name: *msg.Sender.Name,
		},
	}
}
