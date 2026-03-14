package requests

import "github.com/google/uuid"

type FriendRequest struct {
	FriendID uuid.UUID `json:"userId"`
}
