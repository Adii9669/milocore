package dto

import "time"

type MessageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`

	CreatedAt time.Time `json:"createdAt"`

	//for dm
	IsMine *bool `json:"isMine,omitempty"`

	//for crews
	Sender *SenderDTO `json:"sender,omitempty"`
	CrewID *string    `json:"crewId,omitempty"`
}

type SenderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
