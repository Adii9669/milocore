package dto

import "time"

type MessageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`

	CreatedAt time.Time `json:"createdAt"`

	//for dm
	IsMine *bool `json:"isMine,omitempty"`

	//for channels
	Sender         *SenderDTO `json:"sender,omitempty"`
	ChannelID      *string    `json:"channelId,omitempty"`
	FileURL        *string    `json:"fileUrl,omitempty"`
	FilePreviewURL *string    `json:"filePreviewUrl,omitempty"`
	FileName       *string    `json:"fileName,omitempty"`
	FileSize       *int64     `json:"fileSize,omitempty"`
	MimeType       *string    `json:"mimeType,omitempty"`
}

type SenderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
