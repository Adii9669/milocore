package dto

import "time"

type PlanPollResponse struct {
	ID        string                   `json:"id"`
	PlanID    string                   `json:"planId"`
	Question  string                   `json:"question"`
	Options   []PlanPollOptionResponse `json:"options"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}

type PlanPollOptionResponse struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	VoteCount    int    `json:"voteCount"`
	SelectedByMe bool   `json:"selectedByMe"`
}
