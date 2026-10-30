package models

import (
	"time"

	"github.com/google/uuid"
)

// PlanPoll provides a single question poll associated with a plan.
type PlanPoll struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PlanID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex" json:"planId"`
	Question  string           `gorm:"size:240;not null;check:chk_plan_polls_question,length(btrim(question)) > 0" json:"question"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Options   []PlanPollOption `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
	Plan      Plan             `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE" json:"-"`
}

type PlanPollOption struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PollID    uuid.UUID      `gorm:"type:uuid;not null;index;check:chk_plan_poll_options_text,length(btrim(text)) > 0" json:"pollId"`
	Text      string         `gorm:"size:200;not null" json:"text"`
	CreatedAt time.Time      `json:"createdAt"`
	Votes     []PlanPollVote `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE" json:"-"`
}

// A user may select any number of options. The composite primary key prevents
// duplicate votes for an option while allowing multiple selections per poll.
type PlanPollVote struct {
	OptionID  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"optionId"`
	UserID    uuid.UUID      `gorm:"type:uuid;primaryKey" json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	Option    PlanPollOption `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE" json:"-"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
