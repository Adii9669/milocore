package dto

import (
	"time"
)

type FriendResponse struct {
	ID        string    `json:"id"` // friend's USER ID
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
