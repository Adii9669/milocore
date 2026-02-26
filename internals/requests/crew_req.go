package requests

import (
	"time"

	"github.com/google/uuid"
)

type CrewRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type UpdateCrewMember struct {
	MutedTil *time.Time `json:"muted_til"`
}

type GetMembers struct {
	CrewID uuid.UUID `json:"crewid"`
}
