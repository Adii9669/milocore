package mapper

import (
	"chat-server/internals/db/models"
	"chat-server/internals/transport/dto"

	"github.com/google/uuid"
)

func ToDMMessageResponse(msg *models.Message, currentUser uuid.UUID) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	content := ""
	if msg.Content != nil {
		content = *msg.Content
	}
	fileURL, previewURL := "", ""
	if msg.ObjectKey != nil {
		fileURL = "/api/files/" + msg.ID.String() + "/download"
		previewURL = "/api/files/" + msg.ID.String() + "/preview"
	}
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "dm",
		Content:   content,
		IsMine:    &isMine,
		CreatedAt: msg.CreatedAt,
		FileURL:   optionalString(fileURL), FilePreviewURL: optionalString(previewURL), FileName: msg.FileName, FileSize: msg.FileSize, MimeType: msg.MimeType,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID.String(),
			Name: msg.Sender.Name,
		},
	}
}

func ToChannelMessageResponse(msg *models.Message, currentUser uuid.UUID) dto.MessageResponse {
	isMine := msg.SenderID == currentUser
	content := ""
	if msg.Content != nil {
		content = *msg.Content
	}
	fileURL, previewURL := "", ""
	if msg.ObjectKey != nil {
		fileURL = "/api/files/" + msg.ID.String() + "/download"
		previewURL = "/api/files/" + msg.ID.String() + "/preview"
	}
	return dto.MessageResponse{
		ID:        msg.ID.String(),
		Type:      "channel",
		Content:   content,
		CreatedAt: msg.CreatedAt,
		IsMine:    &isMine,
		FileURL:   optionalString(fileURL), FilePreviewURL: optionalString(previewURL), FileName: msg.FileName, FileSize: msg.FileSize, MimeType: msg.MimeType,
		Sender: &dto.SenderDTO{
			ID:   msg.SenderID.String(),
			Name: msg.Sender.Name,
		},
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
