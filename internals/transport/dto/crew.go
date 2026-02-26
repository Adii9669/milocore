package dto

import (
	"time"

	"github.com/google/uuid"
)

type CrewResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}

type MemberResponse struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
}

type CrewMembersResponse struct {
	CrewID      uuid.UUID        `json:"crewID"`
	MemberCount int              `json:"member_count"`
	Members     []MemberResponse `json:"members"`
}
