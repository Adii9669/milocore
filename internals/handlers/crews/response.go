package crews

import (
	"github.com/google/uuid"
	"time"
)

type CreateCrewRequest struct {
	Name string `json:"name" binding:"required"`
}

type CrewResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
}
