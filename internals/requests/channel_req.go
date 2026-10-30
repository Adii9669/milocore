package requests

import (
	"time"

	"github.com/google/uuid"
)

type ChannelRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type UpdateChannelMember struct {
	MutedTil *time.Time `json:"muted_til"`
}

type GetMembers struct {
	ChannelID uuid.UUID `json:"channelId"`
}
