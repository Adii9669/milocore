package dto

import "time"

type PlanResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	CreatorID    string     `json:"creatorId"`
	CreatorName  string     `json:"creatorName,omitempty"`
	ChannelID    *string    `json:"channelId,omitempty"`
	DMUserLowID  *string    `json:"dmUserLowId,omitempty"`
	DMUserHighID *string    `json:"dmUserHighId,omitempty"`
	StartAt      time.Time  `json:"startAt"`
	EndAt        *time.Time `json:"endAt,omitempty"`
	Location     *string    `json:"location,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
