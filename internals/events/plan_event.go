package events

import "github.com/google/uuid"

const (
	PlanCreated        = "plan.created"
	PlanResponseChange = "plan.response_changed"
	PlanUpdated        = "plan.updated"
	PlanConfirmed      = "plan.confirmed"
	PlanCancelled      = "plan.cancelled"
	PlanStatusChanged  = "plan.status_changed"
	PlanPollChanged    = "plan.poll_changed"
)

// PlanEvent contains only identifiers needed by authorized clients to refresh
// data through the authenticated Plan APIs. Recipients are routing metadata and
// are never serialized to clients.
type PlanEvent struct {
	Event        string
	EventID      uuid.UUID
	PlanID       uuid.UUID
	ChannelID    *uuid.UUID
	DMUserLowID  *uuid.UUID
	DMUserHighID *uuid.UUID
	Recipients   []uuid.UUID `json:"-"`
}

type PlanEventPublisher interface {
	PublishPlanEvent(PlanEvent)
}
