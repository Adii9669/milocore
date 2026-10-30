package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/events"
	"chat-server/internals/repository"
	"context"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidPlan           = errors.New("invalid plan")
	ErrPlanChatNotFound      = errors.New("chat not found")
	ErrPlanAccessDenied      = errors.New("not authorized to access this chat")
	ErrPlanNotFound          = errors.New("plan not found")
	ErrPlanCreatorOnly       = errors.New("only the plan creator can manage this plan")
	ErrPlanCancelled         = errors.New("cancelled plans cannot be changed")
	ErrPlanStale             = errors.New("plan changed since it was loaded")
	ErrInvalidPlanStatus     = errors.New("invalid plan status")
	ErrInvalidPlanTransition = errors.New("invalid plan status transition")
	ErrInvalidPlanResponse   = errors.New("invalid plan response")
	ErrPlanNotRespondable    = errors.New("plan is not accepting responses")
	ErrInvalidPlanPoll       = errors.New("invalid plan poll")
	ErrPlanPollExists        = errors.New("a poll already exists for this plan")
	ErrPlanPollNotFound      = errors.New("plan poll not found")
	ErrPlanPollClosed        = errors.New("poll is not accepting votes")
)

type PlanInput struct {
	Title       string
	Description *string
	StartAt     time.Time
	EndAt       *time.Time
	Location    *string
}

type UpdatePlanInput struct {
	Title             string
	Description       *string
	StartAt           time.Time
	EndAt             *time.Time
	Location          *string
	ExpectedUpdatedAt time.Time
}

type PlanService interface {
	CreateChannelPlan(context.Context, uuid.UUID, uuid.UUID, PlanInput) (*models.Plan, error)
	CreateDMPlan(context.Context, uuid.UUID, uuid.UUID, PlanInput) (*models.Plan, error)
	ListChannelPlans(context.Context, uuid.UUID, uuid.UUID) ([]models.Plan, error)
	ListDMPlans(context.Context, uuid.UUID, uuid.UUID) ([]models.Plan, error)
	GetPlan(context.Context, uuid.UUID, uuid.UUID) (*models.Plan, error)
	UpdatePlan(context.Context, uuid.UUID, uuid.UUID, UpdatePlanInput) (*models.Plan, error)
	UpdatePlanStatus(context.Context, uuid.UUID, uuid.UUID, string, time.Time) (*models.Plan, error)
	SetResponse(context.Context, uuid.UUID, uuid.UUID, models.PlanResponseStatus) (*PlanParticipationSummary, error)
	GetParticipants(context.Context, uuid.UUID, uuid.UUID) (*PlanParticipationSummary, error)
	CreatePoll(context.Context, uuid.UUID, uuid.UUID, string, []string) (*PlanPollSummary, error)
	GetPoll(context.Context, uuid.UUID, uuid.UUID) (*PlanPollSummary, error)
	SetPollVotes(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) (*PlanPollSummary, error)
}

type PlanPollOptionSummary struct {
	ID           uuid.UUID
	Text         string
	VoteCount    int
	SelectedByMe bool
}

type PlanPollSummary struct {
	ID        uuid.UUID
	PlanID    uuid.UUID
	Question  string
	Options   []PlanPollOptionSummary
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PlanParticipantEntry struct {
	UserID      uuid.UUID
	Name        string
	Image       *string
	Response    *models.PlanResponseStatus
	RespondedAt *time.Time
}

type PlanParticipationSummary struct {
	PlanID               uuid.UUID
	InCount              int
	MaybeCount           int
	CantMakeItCount      int
	PendingResponseCount int
	Participants         []PlanParticipantEntry
}

type planService struct {
	plans    repository.PlanRepository
	channels repository.ChannelRepository
	friends  repository.FriendRepository
	users    repository.UserRepository
	events   events.PlanEventPublisher
}

func NewPlanService(plans repository.PlanRepository, channels repository.ChannelRepository, friends repository.FriendRepository, users repository.UserRepository, publishers ...events.PlanEventPublisher) PlanService {
	service := &planService{plans: plans, channels: channels, friends: friends, users: users}
	if len(publishers) > 0 {
		service.events = publishers[0]
	}
	return service
}

func validatePlanInput(input PlanInput) (PlanInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || utf8.RuneCountInString(input.Title) > 200 || input.StartAt.IsZero() || !input.StartAt.After(time.Now()) {
		return PlanInput{}, ErrInvalidPlan
	}
	if input.EndAt != nil && !input.EndAt.After(input.StartAt) {
		return PlanInput{}, ErrInvalidPlan
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		if value == "" {
			input.Description = nil
		} else {
			input.Description = &value
		}
	}
	if input.Location != nil {
		value := strings.TrimSpace(*input.Location)
		if value == "" {
			input.Location = nil
		} else if utf8.RuneCountInString(value) > 255 {
			return PlanInput{}, ErrInvalidPlan
		} else {
			input.Location = &value
		}
	}
	return input, nil
}

func (s *planService) CreateChannelPlan(ctx context.Context, creatorID, channelID uuid.UUID, input PlanInput) (*models.Plan, error) {
	input, err := validatePlanInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeChannel(ctx, creatorID, channelID); err != nil {
		return nil, err
	}
	plan := &models.Plan{ID: uuid.New(), Title: input.Title, Description: input.Description, CreatorID: creatorID, ChannelID: &channelID, StartAt: input.StartAt, EndAt: input.EndAt, Location: input.Location, Status: models.PlanStatusOpen}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	if err := s.plans.Create(ctx, plan); err != nil {
		return nil, err
	}
	s.publishPlanEvent(plan, events.PlanCreated, recipients)
	return plan, nil
}

func (s *planService) CreateDMPlan(ctx context.Context, creatorID, otherUserID uuid.UUID, input PlanInput) (*models.Plan, error) {
	input, err := validatePlanInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeDM(ctx, creatorID, otherUserID); err != nil {
		return nil, err
	}
	userLowID, userHighID := normalizePair(creatorID, otherUserID)
	plan := &models.Plan{ID: uuid.New(), Title: input.Title, Description: input.Description, CreatorID: creatorID, DMUserLowID: &userLowID, DMUserHighID: &userHighID, StartAt: input.StartAt, EndAt: input.EndAt, Location: input.Location, Status: models.PlanStatusOpen}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	if err := s.plans.Create(ctx, plan); err != nil {
		return nil, err
	}
	s.publishPlanEvent(plan, events.PlanCreated, recipients)
	return plan, nil
}

func (s *planService) ListChannelPlans(ctx context.Context, userID, channelID uuid.UUID) ([]models.Plan, error) {
	if err := s.authorizeChannel(ctx, userID, channelID); err != nil {
		return nil, err
	}
	return s.plans.ListForChannel(ctx, channelID)
}

func (s *planService) ListDMPlans(ctx context.Context, userID, otherUserID uuid.UUID) ([]models.Plan, error) {
	if err := s.authorizeDM(ctx, userID, otherUserID); err != nil {
		return nil, err
	}
	userLowID, userHighID := normalizePair(userID, otherUserID)
	return s.plans.ListForDM(ctx, userLowID, userHighID)
}

func (s *planService) GetPlan(ctx context.Context, userID, planID uuid.UUID) (*models.Plan, error) {
	return s.authorizePlan(ctx, userID, planID)
}

func (s *planService) UpdatePlan(ctx context.Context, userID, planID uuid.UUID, input UpdatePlanInput) (*models.Plan, error) {
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	if plan.CreatorID != userID {
		return nil, ErrPlanCreatorOnly
	}
	if plan.Status == models.PlanStatusCancelled {
		return nil, ErrPlanCancelled
	}
	if input.ExpectedUpdatedAt.IsZero() || !input.ExpectedUpdatedAt.Equal(plan.UpdatedAt) {
		return nil, ErrPlanStale
	}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	input, err = validatePlanUpdateInput(input, plan)
	if err != nil {
		return nil, err
	}
	if err := s.plans.UpdateFields(ctx, plan.ID, input.ExpectedUpdatedAt, input.Title, input.Description, input.StartAt, input.EndAt, input.Location); err != nil {
		if errors.Is(err, repository.ErrPlanStale) {
			return nil, ErrPlanStale
		}
		return nil, err
	}
	updated, err := s.plans.FindByID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	s.publishPlanEvent(updated, events.PlanUpdated, recipients)
	return updated, nil
}

func (s *planService) UpdatePlanStatus(ctx context.Context, userID, planID uuid.UUID, status string, expectedUpdatedAt time.Time) (*models.Plan, error) {
	if !models.IsValidPlanStatus(status) {
		return nil, ErrInvalidPlanStatus
	}
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	if plan.CreatorID != userID {
		return nil, ErrPlanCreatorOnly
	}
	if plan.Status == models.PlanStatusCancelled {
		return nil, ErrPlanCancelled
	}
	if expectedUpdatedAt.IsZero() || !expectedUpdatedAt.Equal(plan.UpdatedAt) {
		return nil, ErrPlanStale
	}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	if !models.IsValidPlanStatus(plan.Status) {
		return nil, ErrInvalidPlanTransition
	}
	// Any active status can transition to any active status, or to cancelled.
	// Cancelled is terminal and handled above.
	if plan.Status == status {
		return plan, nil
	}
	if err := s.plans.UpdateStatus(ctx, plan.ID, expectedUpdatedAt, status); err != nil {
		if errors.Is(err, repository.ErrPlanStale) {
			return nil, ErrPlanStale
		}
		return nil, err
	}
	updated, err := s.plans.FindByID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	eventType := events.PlanStatusChanged
	if status == models.PlanStatusConfirmed {
		eventType = events.PlanConfirmed
	} else if status == models.PlanStatusCancelled {
		eventType = events.PlanCancelled
	}
	s.publishPlanEvent(updated, eventType, recipients)
	return updated, nil
}

func validatePlanUpdateInput(input UpdatePlanInput, existing *models.Plan) (UpdatePlanInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || utf8.RuneCountInString(input.Title) > 200 || input.StartAt.IsZero() {
		return UpdatePlanInput{}, ErrInvalidPlan
	}
	if !input.StartAt.Equal(existing.StartAt) && !input.StartAt.After(time.Now()) {
		return UpdatePlanInput{}, ErrInvalidPlan
	}
	if input.EndAt != nil && !input.EndAt.After(input.StartAt) {
		return UpdatePlanInput{}, ErrInvalidPlan
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		if value == "" {
			input.Description = nil
		} else {
			input.Description = &value
		}
	}
	if input.Location != nil {
		value := strings.TrimSpace(*input.Location)
		if value == "" {
			input.Location = nil
		} else if utf8.RuneCountInString(value) > 255 {
			return UpdatePlanInput{}, ErrInvalidPlan
		} else {
			input.Location = &value
		}
	}
	return input, nil
}

func (s *planService) authorizeChannel(ctx context.Context, userID, channelID uuid.UUID) error {
	exists, err := s.channels.ExistByID(ctx, channelID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPlanChatNotFound
	}
	member, err := s.channels.IsMember(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrPlanAccessDenied
	}
	return nil
}

func (s *planService) authorizeDM(ctx context.Context, userID, otherUserID uuid.UUID) error {
	if userID == otherUserID {
		return ErrInvalidPlan
	}
	exists, err := s.users.ExistByID(ctx, otherUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPlanChatNotFound
	}
	lowID, highID := normalizePair(userID, otherUserID)
	relation, err := s.friends.GetRelation(ctx, lowID, highID)
	if err != nil {
		return err
	}
	if relation == nil || relation.Status != StatusAccepted {
		return ErrPlanAccessDenied
	}
	return nil
}

func (s *planService) authorizePlan(ctx context.Context, userID uuid.UUID, planID uuid.UUID) (*models.Plan, error) {
	plan, err := s.plans.FindByID(ctx, planID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPlanNotFound
	}
	if err != nil {
		return nil, err
	}
	if plan.ChannelID != nil {
		if err := s.authorizeChannel(ctx, userID, *plan.ChannelID); err != nil {
			return nil, err
		}
		return plan, nil
	}
	if plan.DMUserLowID == nil || plan.DMUserHighID == nil {
		return nil, ErrPlanAccessDenied
	}
	var otherUserID uuid.UUID
	switch userID {
	case *plan.DMUserLowID:
		otherUserID = *plan.DMUserHighID
	case *plan.DMUserHighID:
		otherUserID = *plan.DMUserLowID
	default:
		return nil, ErrPlanAccessDenied
	}
	if err := s.authorizeDM(ctx, userID, otherUserID); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *planService) SetResponse(ctx context.Context, userID, planID uuid.UUID, status models.PlanResponseStatus) (*PlanParticipationSummary, error) {
	if !status.IsValid() {
		return nil, ErrInvalidPlanResponse
	}
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	if plan.Status != models.PlanStatusOpen && plan.Status != models.PlanStatusConfirmed {
		return nil, ErrPlanNotRespondable
	}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	if err := s.plans.UpsertParticipantResponse(ctx, &models.PlanParticipant{PlanID: plan.ID, UserID: userID, Status: status}); err != nil {
		return nil, err
	}
	s.publishPlanEvent(plan, events.PlanResponseChange, recipients)
	return s.participationSummary(ctx, plan)
}

func (s *planService) GetParticipants(ctx context.Context, userID, planID uuid.UUID) (*PlanParticipationSummary, error) {
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	return s.participationSummary(ctx, plan)
}

func (s *planService) CreatePoll(ctx context.Context, userID, planID uuid.UUID, question string, options []string) (*PlanPollSummary, error) {
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	if plan.CreatorID != userID {
		return nil, ErrPlanCreatorOnly
	}
	if plan.Status != models.PlanStatusOpen && plan.Status != models.PlanStatusConfirmed {
		return nil, ErrPlanPollClosed
	}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	question, options, err = validatePlanPoll(question, options)
	if err != nil {
		return nil, err
	}
	existing, err := s.plans.FindPollByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPlanPollExists
	}
	poll := &models.PlanPoll{ID: uuid.New(), PlanID: plan.ID, Question: question, Options: make([]models.PlanPollOption, 0, len(options))}
	for _, option := range options {
		poll.Options = append(poll.Options, models.PlanPollOption{ID: uuid.New(), PollID: poll.ID, Text: option})
	}
	if err := s.plans.CreatePoll(ctx, poll); err != nil {
		if errors.Is(err, repository.ErrPlanPollExists) {
			return nil, ErrPlanPollExists
		}
		return nil, err
	}
	s.publishPlanEvent(plan, events.PlanPollChanged, recipients)
	return summarizePlanPoll(poll, userID), nil
}

func (s *planService) GetPoll(ctx context.Context, userID, planID uuid.UUID) (*PlanPollSummary, error) {
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	poll, err := s.plans.FindPollByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, ErrPlanPollNotFound
	}
	return summarizePlanPoll(poll, userID), nil
}

func (s *planService) SetPollVotes(ctx context.Context, userID, planID uuid.UUID, optionIDs []uuid.UUID) (*PlanPollSummary, error) {
	plan, err := s.authorizePlan(ctx, userID, planID)
	if err != nil {
		return nil, err
	}
	if plan.Status != models.PlanStatusOpen && plan.Status != models.PlanStatusConfirmed {
		return nil, ErrPlanPollClosed
	}
	recipients, err := s.planRecipients(ctx, plan)
	if err != nil {
		return nil, err
	}
	poll, err := s.plans.FindPollByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	if poll == nil {
		return nil, ErrPlanPollNotFound
	}
	allowed := make(map[uuid.UUID]struct{}, len(poll.Options))
	for _, option := range poll.Options {
		allowed[option.ID] = struct{}{}
	}
	selected := make(map[uuid.UUID]struct{}, len(optionIDs))
	for _, optionID := range optionIDs {
		if _, ok := allowed[optionID]; !ok {
			return nil, ErrInvalidPlanPoll
		}
		if _, duplicate := selected[optionID]; duplicate {
			return nil, ErrInvalidPlanPoll
		}
		selected[optionID] = struct{}{}
	}
	if err := s.plans.ReplacePollVotes(ctx, poll.ID, userID, optionIDs); err != nil {
		return nil, err
	}
	s.publishPlanEvent(plan, events.PlanPollChanged, recipients)
	poll, err = s.plans.FindPollByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	return summarizePlanPoll(poll, userID), nil
}

func validatePlanPoll(question string, options []string) (string, []string, error) {
	question = strings.TrimSpace(question)
	if question == "" || utf8.RuneCountInString(question) > 240 || len(options) < 2 || len(options) > 10 {
		return "", nil, ErrInvalidPlanPoll
	}
	clean := make([]string, 0, len(options))
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		option = strings.TrimSpace(option)
		key := strings.ToLower(option)
		if option == "" || utf8.RuneCountInString(option) > 200 {
			return "", nil, ErrInvalidPlanPoll
		}
		if _, exists := seen[key]; exists {
			return "", nil, ErrInvalidPlanPoll
		}
		seen[key] = struct{}{}
		clean = append(clean, option)
	}
	return question, clean, nil
}

func summarizePlanPoll(poll *models.PlanPoll, userID uuid.UUID) *PlanPollSummary {
	summary := &PlanPollSummary{ID: poll.ID, PlanID: poll.PlanID, Question: poll.Question, CreatedAt: poll.CreatedAt, UpdatedAt: poll.UpdatedAt, Options: make([]PlanPollOptionSummary, 0, len(poll.Options))}
	for _, option := range poll.Options {
		entry := PlanPollOptionSummary{ID: option.ID, Text: option.Text, VoteCount: len(option.Votes)}
		for _, vote := range option.Votes {
			if vote.UserID == userID {
				entry.SelectedByMe = true
				break
			}
		}
		summary.Options = append(summary.Options, entry)
	}
	return summary
}

func (s *planService) participationSummary(ctx context.Context, plan *models.Plan) (*PlanParticipationSummary, error) {
	memberIDs := make([]uuid.UUID, 0)
	if plan.ChannelID != nil {
		members, err := s.channels.FindMembersByChannelID(ctx, *plan.ChannelID)
		if err != nil {
			return nil, err
		}
		memberIDs = make([]uuid.UUID, 0, len(members))
		for _, member := range members {
			memberIDs = append(memberIDs, member.UserID)
		}
	} else {
		memberIDs = append(memberIDs, *plan.DMUserLowID, *plan.DMUserHighID)
	}
	users, err := s.users.FindByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	responses, err := s.plans.ListParticipantResponses(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	responseByUser := make(map[uuid.UUID]models.PlanParticipant, len(responses))
	for _, response := range responses {
		responseByUser[response.UserID] = response
	}
	summary := &PlanParticipationSummary{PlanID: plan.ID, Participants: make([]PlanParticipantEntry, 0, len(users))}
	for _, user := range users {
		entry := PlanParticipantEntry{UserID: user.ID, Name: user.Name, Image: user.Image}
		if response, ok := responseByUser[user.ID]; ok {
			status := response.Status
			entry.Response = &status
			respondedAt := response.UpdatedAt
			entry.RespondedAt = &respondedAt
			switch status {
			case models.PlanResponseIn:
				summary.InCount++
			case models.PlanResponseMaybe:
				summary.MaybeCount++
			case models.PlanResponseCantMakeIt:
				summary.CantMakeItCount++
			}
		} else {
			summary.PendingResponseCount++
		}
		summary.Participants = append(summary.Participants, entry)
	}
	sort.Slice(summary.Participants, func(i, j int) bool { return summary.Participants[i].Name < summary.Participants[j].Name })
	return summary, nil
}

func (s *planService) planRecipients(ctx context.Context, plan *models.Plan) ([]uuid.UUID, error) {
	if plan.ChannelID != nil {
		members, err := s.channels.FindMembersByChannelID(ctx, *plan.ChannelID)
		if err != nil {
			return nil, err
		}
		recipients := make([]uuid.UUID, 0, len(members))
		seen := make(map[uuid.UUID]struct{}, len(members))
		for _, member := range members {
			if _, exists := seen[member.UserID]; exists {
				continue
			}
			seen[member.UserID] = struct{}{}
			recipients = append(recipients, member.UserID)
		}
		return recipients, nil
	}
	if plan.DMUserLowID == nil || plan.DMUserHighID == nil {
		return nil, ErrPlanAccessDenied
	}
	return []uuid.UUID{*plan.DMUserLowID, *plan.DMUserHighID}, nil
}

func (s *planService) publishPlanEvent(plan *models.Plan, eventType string, recipients []uuid.UUID) {
	if s.events == nil || plan == nil || len(recipients) == 0 {
		return
	}
	s.events.PublishPlanEvent(events.PlanEvent{
		Event: eventType, EventID: uuid.New(), PlanID: plan.ID, ChannelID: plan.ChannelID,
		DMUserLowID: plan.DMUserLowID, DMUserHighID: plan.DMUserHighID, Recipients: recipients,
	})
}
