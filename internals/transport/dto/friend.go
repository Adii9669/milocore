package dto

import (
	"github.com/google/uuid"
	"time"
)

type FriendResponse struct {
	ID        uuid.UUID `json:"id"` // friend's USER ID
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
