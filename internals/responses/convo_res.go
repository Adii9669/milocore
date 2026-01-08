package responses

import (
	"github.com/google/uuid"
	"time"
)

type ConvoResponse struct {
	ID        uuid.UUID `json:"id"`
	UserOneID uuid.UUID `json:"useroneId"`
	UserTwoID uuid.UUID `json:"usertwoId"`
	CreatedAt time.Time `json:"createdAt"`
}
