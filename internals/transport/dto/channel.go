package dto

import (
	"time"

	"github.com/google/uuid"
)

type ChannelResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"ownerId"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type MemberResponse struct {
	UserID   uuid.UUID `json:"userID"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
}

type ChannelMembersResponse struct {
	ChannelID   uuid.UUID        `json:"channelID"`
	MemberCount int              `json:"member_count"`
	Members     []MemberResponse `json:"members"`
}
