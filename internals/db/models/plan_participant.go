package models

import (
	"time"

	"github.com/google/uuid"
)

type PlanResponseStatus string

const (
	PlanResponseIn         PlanResponseStatus = "in"
	PlanResponseMaybe      PlanResponseStatus = "maybe"
	PlanResponseCantMakeIt PlanResponseStatus = "cant_make_it"
)

func (s PlanResponseStatus) IsValid() bool {
	switch s {
	case PlanResponseIn, PlanResponseMaybe, PlanResponseCantMakeIt:
		return true
	default:
		return false
	}
}

// PlanParticipant stores only explicit responses; users who have not responded
// are represented by their absence from this table.
type PlanParticipant struct {
	PlanID    uuid.UUID          `gorm:"type:uuid;primaryKey" json:"planId"`
	UserID    uuid.UUID          `gorm:"type:uuid;primaryKey" json:"userId"`
	Status    PlanResponseStatus `gorm:"type:varchar(24);not null;check:chk_plan_participant_status,status = 'in' OR status = 'maybe' OR status = 'cant_make_it'" json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Plan      Plan               `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"-"`
	User      User               `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
