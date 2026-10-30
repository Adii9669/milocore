package dto

import (
	"time"

	"chat-server/internals/db/models"
)

type PlanParticipantResponse struct {
	UserID      string                     `json:"userId"`
	Name        string                     `json:"name"`
	Image       *string                    `json:"image,omitempty"`
	Response    *models.PlanResponseStatus `json:"response"`
	RespondedAt *time.Time                 `json:"respondedAt,omitempty"`
}

type PlanParticipationSummary struct {
	PlanID               string                    `json:"planId"`
	InCount              int                       `json:"inCount"`
	MaybeCount           int                       `json:"maybeCount"`
	CantMakeItCount      int                       `json:"cantMakeItCount"`
	PendingResponseCount int                       `json:"pendingResponseCount"`
	Participants         []PlanParticipantResponse `json:"participants"`
}
