package repository

import (
	"chat-server/internals/db/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPlanStale = errors.New("plan changed since it was loaded")
var ErrPlanPollExists = errors.New("a poll already exists for this plan")

type PlanRepository interface {
	Create(context.Context, *models.Plan) error
	FindByID(context.Context, uuid.UUID) (*models.Plan, error)
	UpdateFields(context.Context, uuid.UUID, time.Time, string, *string, time.Time, *time.Time, *string) error
	UpdateStatus(context.Context, uuid.UUID, time.Time, string) error
	ListForChannel(context.Context, uuid.UUID) ([]models.Plan, error)
	ListForDM(context.Context, uuid.UUID, uuid.UUID) ([]models.Plan, error)
	UpsertParticipantResponse(context.Context, *models.PlanParticipant) error
	ListParticipantResponses(context.Context, uuid.UUID) ([]models.PlanParticipant, error)
	CreatePoll(context.Context, *models.PlanPoll) error
	FindPollByPlanID(context.Context, uuid.UUID) (*models.PlanPoll, error)
	ReplacePollVotes(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error
}

type planRepository struct{ db *gorm.DB }

func NewPlanRepository(db *gorm.DB) PlanRepository { return &planRepository{db: db} }

func (r *planRepository) Create(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) FindByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	var plan models.Plan
	if err := r.db.WithContext(ctx).Preload("Creator").First(&plan, "id = ?", planID).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) UpdateFields(ctx context.Context, planID uuid.UUID, expectedUpdatedAt time.Time, title string, description *string, startAt time.Time, endAt *time.Time, location *string) error {
	result := r.db.WithContext(ctx).Model(&models.Plan{}).
		Where("id = ? AND updated_at = ?", planID, expectedUpdatedAt).
		Updates(map[string]any{
			"title": title, "description": description, "start_at": startAt,
			"end_at": endAt, "location": location, "updated_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPlanStale
	}
	return nil
}

func (r *planRepository) UpdateStatus(ctx context.Context, planID uuid.UUID, expectedUpdatedAt time.Time, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Plan{}).
		Where("id = ? AND updated_at = ?", planID, expectedUpdatedAt).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPlanStale
	}
	return nil
}

func (r *planRepository) UpsertParticipantResponse(ctx context.Context, response *models.PlanParticipant) error {
	response.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plan_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
	}).Create(response).Error
}

func (r *planRepository) ListParticipantResponses(ctx context.Context, planID uuid.UUID) ([]models.PlanParticipant, error) {
	var responses []models.PlanParticipant
	err := r.db.WithContext(ctx).Where("plan_id = ?", planID).Find(&responses).Error
	return responses, err
}

func (r *planRepository) CreatePoll(ctx context.Context, poll *models.PlanPoll) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "plan_id"}},
			DoNothing: true,
		}).Omit("Options").Create(poll)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrPlanPollExists
		}
		for i := range poll.Options {
			poll.Options[i].PollID = poll.ID
		}
		return tx.Create(&poll.Options).Error
	})
}

func (r *planRepository) FindPollByPlanID(ctx context.Context, planID uuid.UUID) (*models.PlanPoll, error) {
	var poll models.PlanPoll
	err := r.db.WithContext(ctx).
		Preload("Options", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC, id ASC") }).
		Preload("Options.Votes").
		First(&poll, "plan_id = ?", planID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &poll, nil
}

func (r *planRepository) ReplacePollVotes(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var poll models.PlanPoll
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&poll, "id = ?", pollID).Error; err != nil {
			return err
		}
		optionsForPoll := tx.Model(&models.PlanPollOption{}).Select("id").Where("poll_id = ?", pollID)
		if err := tx.Where("user_id = ? AND option_id IN (?)", userID, optionsForPoll).Delete(&models.PlanPollVote{}).Error; err != nil {
			return err
		}
		votes := make([]models.PlanPollVote, 0, len(optionIDs))
		for _, optionID := range optionIDs {
			votes = append(votes, models.PlanPollVote{OptionID: optionID, UserID: userID})
		}
		if len(votes) == 0 {
			return nil
		}
		return tx.Create(&votes).Error
	})
}

func (r *planRepository) ListForChannel(ctx context.Context, channelID uuid.UUID) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.WithContext(ctx).Where("channel_id = ?", channelID).Order("start_at ASC, created_at ASC").Find(&plans).Error
	return plans, err
}

func (r *planRepository) ListForDM(ctx context.Context, userLowID, userHighID uuid.UUID) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.WithContext(ctx).Where("dm_user_low_id = ? AND dm_user_high_id = ?", userLowID, userHighID).Order("start_at ASC, created_at ASC").Find(&plans).Error
	return plans, err
}
