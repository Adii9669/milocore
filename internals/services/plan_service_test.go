package services

import (
	"chat-server/internals/db/models"
	"chat-server/internals/events"
	"chat-server/internals/repository"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRepoStub struct {
	created         *models.Plan
	plan            *models.Plan
	channelPlans    []models.Plan
	dmPlans         []models.Plan
	responses       map[uuid.UUID]models.PlanParticipant
	poll            *models.PlanPoll
	listedChannelID uuid.UUID
	listedLowID     uuid.UUID
	listedHighID    uuid.UUID
}

type planPublisherStub struct{ published []events.PlanEvent }

func (p *planPublisherStub) PublishPlanEvent(event events.PlanEvent) {
	p.published = append(p.published, event)
}

func (r *planRepoStub) FindByID(_ context.Context, id uuid.UUID) (*models.Plan, error) {
	if r.plan == nil || r.plan.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.plan
	return &copy, nil
}
func (r *planRepoStub) UpdateFields(_ context.Context, id uuid.UUID, expected time.Time, title string, description *string, start time.Time, end *time.Time, location *string) error {
	if r.plan == nil || r.plan.ID != id || !r.plan.UpdatedAt.Equal(expected) {
		return repository.ErrPlanStale
	}
	r.plan.Title, r.plan.Description, r.plan.StartAt, r.plan.EndAt, r.plan.Location = title, description, start, end, location
	r.plan.UpdatedAt = expected.Add(time.Second)
	return nil
}
func (r *planRepoStub) UpdateStatus(_ context.Context, id uuid.UUID, expected time.Time, status string) error {
	if r.plan == nil || r.plan.ID != id || !r.plan.UpdatedAt.Equal(expected) {
		return repository.ErrPlanStale
	}
	r.plan.Status = status
	r.plan.UpdatedAt = expected.Add(time.Second)
	return nil
}
func (r *planRepoStub) UpsertParticipantResponse(_ context.Context, response *models.PlanParticipant) error {
	if r.responses == nil {
		r.responses = make(map[uuid.UUID]models.PlanParticipant)
	}
	r.responses[response.UserID] = *response
	return nil
}
func (r *planRepoStub) ListParticipantResponses(_ context.Context, planID uuid.UUID) ([]models.PlanParticipant, error) {
	result := make([]models.PlanParticipant, 0, len(r.responses))
	for _, response := range r.responses {
		if response.PlanID == planID {
			result = append(result, response)
		}
	}
	return result, nil
}

func (r *planRepoStub) CreatePoll(_ context.Context, poll *models.PlanPoll) error {
	if r.poll != nil {
		return repository.ErrPlanPollExists
	}
	copy := *poll
	copy.Options = append([]models.PlanPollOption(nil), poll.Options...)
	r.poll = &copy
	return nil
}
func (r *planRepoStub) FindPollByPlanID(_ context.Context, planID uuid.UUID) (*models.PlanPoll, error) {
	if r.poll == nil || r.poll.PlanID != planID {
		return nil, nil
	}
	copy := *r.poll
	copy.Options = append([]models.PlanPollOption(nil), r.poll.Options...)
	for i := range copy.Options {
		copy.Options[i].Votes = append([]models.PlanPollVote(nil), r.poll.Options[i].Votes...)
	}
	return &copy, nil
}
func (r *planRepoStub) ReplacePollVotes(_ context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	if r.poll == nil || r.poll.ID != pollID {
		return gorm.ErrRecordNotFound
	}
	for i := range r.poll.Options {
		votes := r.poll.Options[i].Votes[:0]
		for _, vote := range r.poll.Options[i].Votes {
			if vote.UserID != userID {
				votes = append(votes, vote)
			}
		}
		r.poll.Options[i].Votes = votes
		for _, optionID := range optionIDs {
			if optionID == r.poll.Options[i].ID {
				r.poll.Options[i].Votes = append(r.poll.Options[i].Votes, models.PlanPollVote{OptionID: optionID, UserID: userID})
			}
		}
	}
	return nil
}

func (r *planRepoStub) Create(_ context.Context, plan *models.Plan) error {
	copy := *plan
	r.created = &copy
	return nil
}
func (r *planRepoStub) ListForChannel(_ context.Context, channelID uuid.UUID) ([]models.Plan, error) {
	r.listedChannelID = channelID
	return r.channelPlans, nil
}
func (r *planRepoStub) ListForDM(_ context.Context, userLowID, userHighID uuid.UUID) ([]models.Plan, error) {
	r.listedLowID, r.listedHighID = userLowID, userHighID
	return r.dmPlans, nil
}

type planChannelRepoStub struct {
	repository.ChannelRepository
	exists, member bool
	members        []models.ChannelMember
}

func (r planChannelRepoStub) FindMembersByChannelID(context.Context, uuid.UUID) ([]models.ChannelMember, error) {
	return r.members, nil
}

func (r planChannelRepoStub) ExistByID(context.Context, uuid.UUID) (bool, error) {
	return r.exists, nil
}
func (r planChannelRepoStub) IsMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return r.member, nil
}

type planFriendRepoStub struct {
	repository.FriendRepository
	relation *models.Friend
}

func (r planFriendRepoStub) GetRelation(context.Context, uuid.UUID, uuid.UUID) (*models.Friend, error) {
	return r.relation, nil
}

type planUserRepoStub struct {
	repository.UserRepository
	exists  bool
	records []models.User
}

func (r planUserRepoStub) ExistByID(context.Context, uuid.UUID) (bool, error) { return r.exists, nil }
func (r planUserRepoStub) FindByIDs(_ context.Context, ids []uuid.UUID) ([]models.User, error) {
	wanted := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	result := make([]models.User, 0, len(ids))
	for _, user := range r.records {
		if wanted[user.ID] {
			result = append(result, user)
		}
	}
	return result, nil
}

func TestCreateChannelPlanPersistsCreatorAndChannel(t *testing.T) {
	creatorID, channelID := uuid.New(), uuid.New()
	repo := &planRepoStub{}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	start := time.Now().Add(2 * time.Hour)
	plan, err := service.CreateChannelPlan(context.Background(), creatorID, channelID, PlanInput{Title: "  Dinner  ", StartAt: start})
	if err != nil {
		t.Fatalf("CreateChannelPlan returned error: %v", err)
	}
	if plan.Title != "Dinner" || plan.CreatorID != creatorID || plan.ChannelID == nil || *plan.ChannelID != channelID || plan.Status != models.PlanStatusOpen {
		t.Fatalf("unexpected plan result: %#v", plan)
	}
	if repo.created == nil || repo.created.ID != plan.ID {
		t.Fatal("plan was not passed to repository")
	}
}

func TestCreateChannelPlanRejectsInvalidInput(t *testing.T) {
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	validStart := time.Now().Add(time.Hour)
	for _, input := range []PlanInput{
		{Title: "  ", StartAt: validStart},
		{Title: "Plan", StartAt: time.Time{}},
		{Title: "Plan", StartAt: validStart, EndAt: &validStart},
		{Title: "Plan", StartAt: time.Now().Add(-time.Minute)},
	} {
		if _, err := service.CreateChannelPlan(context.Background(), uuid.New(), uuid.New(), input); !errors.Is(err, ErrInvalidPlan) {
			t.Errorf("expected ErrInvalidPlan, got %v", err)
		}
	}
}

func TestCreateChannelPlanRequiresExistingChannelMembership(t *testing.T) {
	creatorID, channelID := uuid.New(), uuid.New()
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{exists: true, member: false}, planFriendRepoStub{}, planUserRepoStub{})
	_, err := service.CreateChannelPlan(context.Background(), creatorID, channelID, PlanInput{Title: "Plan", StartAt: time.Now().Add(time.Hour)})
	if !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}

	service = NewPlanService(&planRepoStub{}, planChannelRepoStub{exists: false}, planFriendRepoStub{}, planUserRepoStub{})
	_, err = service.CreateChannelPlan(context.Background(), creatorID, channelID, PlanInput{Title: "Plan", StartAt: time.Now().Add(time.Hour)})
	if !errors.Is(err, ErrPlanChatNotFound) {
		t.Fatalf("expected chat not found, got %v", err)
	}
}

func TestCreateDMPlanUsesCanonicalAcceptedFriendPair(t *testing.T) {
	creatorID, otherID := uuid.New(), uuid.New()
	low, high := normalizePair(creatorID, otherID)
	repo := &planRepoStub{}
	service := NewPlanService(repo, planChannelRepoStub{}, planFriendRepoStub{relation: &models.Friend{Status: StatusAccepted}}, planUserRepoStub{exists: true})
	plan, err := service.CreateDMPlan(context.Background(), creatorID, otherID, PlanInput{Title: "Coffee", StartAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("CreateDMPlan returned error: %v", err)
	}
	if plan.DMUserLowID == nil || *plan.DMUserLowID != low || plan.DMUserHighID == nil || *plan.DMUserHighID != high || plan.CreatorID != creatorID {
		t.Fatalf("unexpected DM association: %#v", plan)
	}
}

func TestCreateDMPlanRejectsNonFriend(t *testing.T) {
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{}, planFriendRepoStub{}, planUserRepoStub{exists: true})
	_, err := service.CreateDMPlan(context.Background(), uuid.New(), uuid.New(), PlanInput{Title: "Coffee", StartAt: time.Now().Add(time.Hour)})
	if !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

func TestListPlansRequiresConversationAccess(t *testing.T) {
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{exists: true, member: false}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := service.ListChannelPlans(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

func TestListDMPlansUsesTheAuthorizedCanonicalChat(t *testing.T) {
	userID, otherID := uuid.New(), uuid.New()
	low, high := normalizePair(userID, otherID)
	want := []models.Plan{{ID: uuid.New(), Title: "Dinner"}}
	repo := &planRepoStub{dmPlans: want}
	service := NewPlanService(repo, planChannelRepoStub{}, planFriendRepoStub{relation: &models.Friend{Status: StatusAccepted}}, planUserRepoStub{exists: true})
	got, err := service.ListDMPlans(context.Background(), userID, otherID)
	if err != nil {
		t.Fatalf("ListDMPlans returned error: %v", err)
	}
	if repo.listedLowID != low || repo.listedHighID != high {
		t.Fatalf("queried wrong DM pair: %s / %s", repo.listedLowID, repo.listedHighID)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("unexpected plans: %#v", got)
	}
}

func TestListDMPlansRejectsNonFriend(t *testing.T) {
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{}, planFriendRepoStub{}, planUserRepoStub{exists: true})
	if _, err := service.ListDMPlans(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

func TestCreateDMPlanReturnsNotFoundForMissingUser(t *testing.T) {
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{}, planFriendRepoStub{}, planUserRepoStub{exists: false})
	_, err := service.CreateDMPlan(context.Background(), uuid.New(), uuid.New(), PlanInput{Title: "Plan", StartAt: time.Now().Add(time.Hour)})
	if !errors.Is(err, ErrPlanChatNotFound) {
		t.Fatalf("expected chat not found, got %v", err)
	}
}

func TestUpdatePlanEditsCreatorFieldsWithoutChangingUneditedSchedule(t *testing.T) {
	userID, channelID, planID := uuid.New(), uuid.New(), uuid.New()
	startAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	updatedAt := time.Now().Add(-time.Minute).Truncate(time.Second)
	plan := &models.Plan{ID: planID, CreatorID: userID, ChannelID: &channelID, Title: "Before", StartAt: startAt, UpdatedAt: updatedAt, Status: models.PlanStatusOpen}
	participantID := uuid.New()
	repo := &planRepoStub{plan: plan, responses: map[uuid.UUID]models.PlanParticipant{
		participantID: {PlanID: planID, UserID: participantID, Status: models.PlanResponseIn},
	}}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	description := "New details"
	updated, err := service.UpdatePlan(context.Background(), userID, planID, UpdatePlanInput{
		Title: " After ", Description: &description, StartAt: startAt, ExpectedUpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("UpdatePlan returned error: %v", err)
	}
	if updated.Title != "After" || updated.Description == nil || *updated.Description != description || !updated.StartAt.Equal(startAt) {
		t.Fatalf("unexpected updated plan: %#v", updated)
	}
	if !updated.UpdatedAt.After(updatedAt) {
		t.Fatalf("update timestamp did not advance: before=%s after=%s", updatedAt, updated.UpdatedAt)
	}
	if len(repo.responses) != 1 || repo.responses[participantID].Status != models.PlanResponseIn {
		t.Fatalf("plan edit changed participant responses: %#v", repo.responses)
	}
}

func TestUpdatePlanRejectsNonCreatorStaleAndCancelledRequests(t *testing.T) {
	userID, creatorID, channelID, planID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	updatedAt := time.Now().Truncate(time.Second)
	startAt := time.Now().Add(time.Hour)
	input := UpdatePlanInput{Title: "Changed", StartAt: startAt, ExpectedUpdatedAt: updatedAt}
	newService := func(plan *models.Plan) (*planRepoStub, PlanService) {
		repo := &planRepoStub{plan: plan}
		return repo, NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	}

	_, service := newService(&models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, UpdatedAt: updatedAt, Status: models.PlanStatusOpen})
	if _, err := service.UpdatePlan(context.Background(), userID, planID, input); !errors.Is(err, ErrPlanCreatorOnly) {
		t.Fatalf("expected creator-only error, got %v", err)
	}

	repo, service := newService(&models.Plan{ID: planID, CreatorID: userID, ChannelID: &channelID, UpdatedAt: updatedAt, Status: models.PlanStatusOpen})
	input.ExpectedUpdatedAt = updatedAt.Add(-time.Second)
	if _, err := service.UpdatePlan(context.Background(), userID, planID, input); !errors.Is(err, ErrPlanStale) {
		t.Fatalf("expected stale plan error, got %v", err)
	}
	if repo.plan.Title != "" {
		t.Fatalf("stale update modified plan: %#v", repo.plan)
	}

	input.ExpectedUpdatedAt = updatedAt
	_, service = newService(&models.Plan{ID: planID, CreatorID: userID, ChannelID: &channelID, UpdatedAt: updatedAt, Status: models.PlanStatusCancelled})
	if _, err := service.UpdatePlan(context.Background(), userID, planID, input); !errors.Is(err, ErrPlanCancelled) {
		t.Fatalf("expected cancelled plan error, got %v", err)
	}
}

func TestPlanUpdateRejectsInvalidFieldsAndPastScheduleChanges(t *testing.T) {
	existingStart := time.Now().Add(time.Hour)
	existing := &models.Plan{StartAt: existingStart}
	pastStart := time.Now().Add(-time.Hour)
	tooLongLocation := strings.Repeat("x", 256)
	badEnd := existingStart.Add(-time.Minute)
	for _, input := range []UpdatePlanInput{
		{Title: " ", StartAt: existingStart},
		{Title: "Valid", StartAt: time.Time{}},
		{Title: "Valid", StartAt: existingStart, EndAt: &badEnd},
		{Title: "Valid", StartAt: pastStart},
		{Title: "Valid", StartAt: existingStart, Location: &tooLongLocation},
	} {
		if _, err := validatePlanUpdateInput(input, existing); !errors.Is(err, ErrInvalidPlan) {
			t.Errorf("expected invalid plan update, got %v for %#v", err, input)
		}
	}
}

func TestPlanStatusFlexibleTransitionsAndCancellationPreserveResponses(t *testing.T) {
	activeStatuses := []string{models.PlanStatusOpen, models.PlanStatusConfirmed, models.PlanStatusCompleted}
	targets := append(append([]string{}, activeStatuses...), models.PlanStatusCancelled)
	for _, current := range activeStatuses {
		for _, target := range targets {
			t.Run(current+"_to_"+target, func(t *testing.T) {
				creatorID, channelID, planID, participantID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				updatedAt := time.Now().Add(-time.Minute).Truncate(time.Second)
				plan := &models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, Status: current, UpdatedAt: updatedAt}
				repo := &planRepoStub{plan: plan, responses: map[uuid.UUID]models.PlanParticipant{
					participantID: {PlanID: planID, UserID: participantID, Status: models.PlanResponseMaybe},
				}}
				service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
				updated, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, target, updatedAt)
				if err != nil {
					t.Fatalf("UpdatePlanStatus returned error: %v", err)
				}
				if updated.Status != target {
					t.Fatalf("status = %q, want %q", updated.Status, target)
				}
				if len(repo.responses) != 1 || repo.responses[participantID].Status != models.PlanResponseMaybe {
					t.Fatalf("status change modified participation responses: %#v", repo.responses)
				}
				if current == target && !updated.UpdatedAt.Equal(updatedAt) {
					t.Fatal("same-status request should be a no-op")
				}
				if current != target && !updated.UpdatedAt.After(updatedAt) {
					t.Fatal("status change did not advance update timestamp")
				}
			})
		}
	}
}

func TestCancelledPlanIsTerminalAndStatusChangeRequiresCreator(t *testing.T) {
	creatorID, otherMemberID, channelID, planID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	updatedAt := time.Now().Add(-time.Minute).Truncate(time.Second)
	repo := &planRepoStub{plan: &models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, Status: models.PlanStatusCancelled, UpdatedAt: updatedAt}}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, models.PlanStatusOpen, updatedAt); !errors.Is(err, ErrPlanCancelled) {
		t.Fatalf("expected cancelled plan to reject restoration, got %v", err)
	}

	repo.plan.Status = models.PlanStatusOpen
	if _, err := service.UpdatePlanStatus(context.Background(), otherMemberID, planID, models.PlanStatusConfirmed, updatedAt); !errors.Is(err, ErrPlanCreatorOnly) {
		t.Fatalf("expected non-creator to be denied, got %v", err)
	}
	service = NewPlanService(repo, planChannelRepoStub{exists: true, member: false}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, models.PlanStatusConfirmed, updatedAt); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected non-member to be denied, got %v", err)
	}

	repo.plan.Status = models.PlanStatusOpen
	service = NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, models.PlanStatusConfirmed, updatedAt.Add(-time.Second)); !errors.Is(err, ErrPlanStale) {
		t.Fatalf("expected stale status update to be rejected, got %v", err)
	}
	if _, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, "invalid", updatedAt); !errors.Is(err, ErrInvalidPlanStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestPlanResponseTransitionsUpdateOneRecordAndCountPendingMembers(t *testing.T) {
	planID, channelID := uuid.New(), uuid.New()
	userA, userB := uuid.New(), uuid.New()
	plan := &models.Plan{ID: planID, ChannelID: &channelID, Status: models.PlanStatusOpen}
	repo := &planRepoStub{plan: plan}
	channels := planChannelRepoStub{exists: true, member: true, members: []models.ChannelMember{{UserID: userA}, {UserID: userB}}}
	users := planUserRepoStub{records: []models.User{{ID: userA, Name: "A"}, {ID: userB, Name: "B"}}}
	service := NewPlanService(repo, channels, planFriendRepoStub{}, users)

	first, err := service.SetResponse(context.Background(), userA, planID, models.PlanResponseIn)
	if err != nil {
		t.Fatalf("SetResponse(in) returned error: %v", err)
	}
	if first.InCount != 1 || first.PendingResponseCount != 1 || first.CantMakeItCount != 0 {
		t.Fatalf("unexpected first summary: %#v", first)
	}
	second, err := service.SetResponse(context.Background(), userA, planID, models.PlanResponseMaybe)
	if err != nil {
		t.Fatalf("SetResponse(maybe) returned error: %v", err)
	}
	if len(repo.responses) != 1 || second.MaybeCount != 1 || second.InCount != 0 || second.PendingResponseCount != 1 {
		t.Fatalf("response transition created duplicate or incorrect counts: rows=%d summary=%#v", len(repo.responses), second)
	}
	if second.Participants[0].Response == nil && second.Participants[1].Response == nil {
		t.Fatal("responding participant was not assigned a response")
	}
}

func TestPlanResponseSupportsEveryTransitionAndRepeatedSubmissions(t *testing.T) {
	statuses := []models.PlanResponseStatus{models.PlanResponseIn, models.PlanResponseMaybe, models.PlanResponseCantMakeIt}
	for _, previous := range statuses {
		for _, next := range statuses {
			t.Run(string(previous)+"_to_"+string(next), func(t *testing.T) {
				planID, channelID, userID, pendingID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
				repo := &planRepoStub{
					plan: &models.Plan{ID: planID, ChannelID: &channelID, Status: models.PlanStatusOpen},
					responses: map[uuid.UUID]models.PlanParticipant{
						userID: {PlanID: planID, UserID: userID, Status: previous},
					},
				}
				channelRepo := planChannelRepoStub{exists: true, member: true, members: []models.ChannelMember{{UserID: userID}, {UserID: pendingID}}}
				userRepo := planUserRepoStub{records: []models.User{{ID: userID, Name: "A"}, {ID: pendingID, Name: "B"}}}
				service := NewPlanService(repo, channelRepo, planFriendRepoStub{}, userRepo)
				summary, err := service.SetResponse(context.Background(), userID, planID, next)
				if err != nil {
					t.Fatalf("SetResponse returned error: %v", err)
				}
				if len(repo.responses) != 1 || summary.PendingResponseCount != 1 {
					t.Fatalf("expected one response and one pending member, got %d rows and %#v", len(repo.responses), summary)
				}
				if summary.InCount != countStatus(next, models.PlanResponseIn) ||
					summary.MaybeCount != countStatus(next, models.PlanResponseMaybe) ||
					summary.CantMakeItCount != countStatus(next, models.PlanResponseCantMakeIt) {
					t.Fatalf("incorrect counts after transition: %#v", summary)
				}
				if _, err := service.SetResponse(context.Background(), userID, planID, next); err != nil {
					t.Fatalf("repeated SetResponse returned error: %v", err)
				}
				if len(repo.responses) != 1 {
					t.Fatalf("repeated request created duplicate row: %d", len(repo.responses))
				}
			})
		}
	}
}

func countStatus(actual, expected models.PlanResponseStatus) int {
	if actual == expected {
		return 1
	}
	return 0
}

func TestPlanResponseRejectsInvalidOrUnauthorizedAndListsOnlyCurrentMembers(t *testing.T) {
	planID, channelID := uuid.New(), uuid.New()
	memberID, outsiderID, formerID := uuid.New(), uuid.New(), uuid.New()
	repo := &planRepoStub{
		plan: &models.Plan{ID: planID, ChannelID: &channelID, Status: models.PlanStatusOpen},
		responses: map[uuid.UUID]models.PlanParticipant{
			formerID: {PlanID: planID, UserID: formerID, Status: models.PlanResponseIn},
		},
	}
	channels := planChannelRepoStub{exists: true, member: true, members: []models.ChannelMember{{UserID: memberID}}}
	users := planUserRepoStub{records: []models.User{{ID: memberID, Name: "Member"}, {ID: formerID, Name: "Former"}}}
	service := NewPlanService(repo, channels, planFriendRepoStub{}, users)
	if _, err := service.SetResponse(context.Background(), memberID, planID, "reject"); !errors.Is(err, ErrInvalidPlanResponse) {
		t.Fatalf("expected invalid response error, got %v", err)
	}
	unauthorizedService := NewPlanService(repo, planChannelRepoStub{exists: true, member: false, members: channels.members}, planFriendRepoStub{}, users)
	if _, err := unauthorizedService.SetResponse(context.Background(), outsiderID, planID, models.PlanResponseIn); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
	summary, err := service.GetParticipants(context.Background(), memberID, planID)
	if err != nil {
		t.Fatalf("GetParticipants returned error: %v", err)
	}
	if len(summary.Participants) != 1 || summary.Participants[0].UserID != memberID || summary.InCount != 0 || summary.PendingResponseCount != 1 {
		t.Fatalf("former member response should not appear or count: %#v", summary)
	}
}

func TestCanceledPlanRejectsNewResponses(t *testing.T) {
	planID, channelID, userID := uuid.New(), uuid.New(), uuid.New()
	repo := &planRepoStub{plan: &models.Plan{ID: planID, ChannelID: &channelID, Status: "cancelled"}}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := service.SetResponse(context.Background(), userID, planID, models.PlanResponseIn); !errors.Is(err, ErrPlanNotRespondable) {
		t.Fatalf("expected non-respondable plan error, got %v", err)
	}
}

func TestDMPlanParticipationRequiresAcceptedPairAndIncludesBothUsers(t *testing.T) {
	userID, friendID, outsiderID := uuid.New(), uuid.New(), uuid.New()
	lowID, highID := normalizePair(userID, friendID)
	planID := uuid.New()
	repo := &planRepoStub{plan: &models.Plan{
		ID: planID, DMUserLowID: &lowID, DMUserHighID: &highID, Status: models.PlanStatusOpen,
	}}
	users := planUserRepoStub{exists: true, records: []models.User{{ID: userID, Name: "A"}, {ID: friendID, Name: "B"}}}
	service := NewPlanService(repo, planChannelRepoStub{}, planFriendRepoStub{relation: &models.Friend{Status: StatusAccepted}}, users)
	summary, err := service.SetResponse(context.Background(), userID, planID, models.PlanResponseIn)
	if err != nil {
		t.Fatalf("SetResponse returned error: %v", err)
	}
	if summary.InCount != 1 || summary.PendingResponseCount != 1 || len(summary.Participants) != 2 {
		t.Fatalf("unexpected DM participation summary: %#v", summary)
	}
	unauthorizedService := NewPlanService(repo, planChannelRepoStub{}, planFriendRepoStub{relation: &models.Friend{Status: StatusAccepted}}, users)
	if _, err := unauthorizedService.GetParticipants(context.Background(), outsiderID, planID); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected outsider to be denied, got %v", err)
	}
}

func TestPlanPollSupportsMultipleSelectionsAndReplaceableVotes(t *testing.T) {
	planID, channelID, creatorID, memberID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	startAt := time.Now().Add(48 * time.Hour)
	repo := &planRepoStub{plan: &models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, Status: models.PlanStatusOpen, StartAt: startAt}}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	poll, err := service.CreatePoll(context.Background(), creatorID, planID, "Which time works?", []string{"Saturday", "Sunday"})
	if err != nil {
		t.Fatalf("CreatePoll returned error: %v", err)
	}
	if poll.Question != "Which time works?" || len(poll.Options) != 2 {
		t.Fatalf("unexpected poll: %#v", poll)
	}
	optionIDs := []uuid.UUID{poll.Options[0].ID, poll.Options[1].ID}
	poll, err = service.SetPollVotes(context.Background(), memberID, planID, optionIDs)
	if err != nil {
		t.Fatalf("SetPollVotes returned error: %v", err)
	}
	for _, option := range poll.Options {
		if !option.SelectedByMe || option.VoteCount != 1 {
			t.Fatalf("multi-select vote was not counted: %#v", option)
		}
	}
	poll, err = service.SetPollVotes(context.Background(), memberID, planID, optionIDs[:1])
	if err != nil {
		t.Fatalf("changing votes returned error: %v", err)
	}
	if poll.Options[0].VoteCount != 1 || !poll.Options[0].SelectedByMe || poll.Options[1].VoteCount != 0 || poll.Options[1].SelectedByMe {
		t.Fatalf("replaced votes were not reflected: %#v", poll.Options)
	}
	if !repo.plan.StartAt.Equal(startAt) {
		t.Fatal("poll votes changed the plan date/time")
	}
}

func TestPlanPollAuthorizationValidationAndClosedPlan(t *testing.T) {
	planID, channelID, creatorID, outsiderID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &planRepoStub{plan: &models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, Status: models.PlanStatusOpen}}
	memberService := NewPlanService(repo, planChannelRepoStub{exists: true, member: true}, planFriendRepoStub{}, planUserRepoStub{})
	for _, input := range []struct {
		question string
		options  []string
	}{
		{"", []string{"A", "B"}}, {"Question", []string{"A"}}, {"Question", []string{"A", " a "}},
	} {
		if _, err := memberService.CreatePoll(context.Background(), creatorID, planID, input.question, input.options); !errors.Is(err, ErrInvalidPlanPoll) {
			t.Errorf("expected invalid poll input error, got %v", err)
		}
	}
	unauthorized := NewPlanService(repo, planChannelRepoStub{exists: true, member: false}, planFriendRepoStub{}, planUserRepoStub{})
	if _, err := unauthorized.GetPoll(context.Background(), outsiderID, planID); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected unauthorized member denied, got %v", err)
	}
	poll, err := memberService.CreatePoll(context.Background(), creatorID, planID, "Pick", []string{"A", "B"})
	if err != nil {
		t.Fatalf("CreatePoll returned error: %v", err)
	}
	if _, err := memberService.CreatePoll(context.Background(), creatorID, planID, "Second", []string{"C", "D"}); !errors.Is(err, ErrPlanPollExists) {
		t.Fatalf("expected one poll per plan, got %v", err)
	}
	if _, err := unauthorized.SetPollVotes(context.Background(), outsiderID, planID, []uuid.UUID{poll.Options[0].ID}); !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected unauthorized vote denied, got %v", err)
	}
	if _, err := memberService.SetPollVotes(context.Background(), creatorID, planID, []uuid.UUID{uuid.New()}); !errors.Is(err, ErrInvalidPlanPoll) {
		t.Fatalf("expected foreign option rejected, got %v", err)
	}
	repo.plan.Status = models.PlanStatusCancelled
	if _, err := memberService.SetPollVotes(context.Background(), creatorID, planID, []uuid.UUID{poll.Options[0].ID}); !errors.Is(err, ErrPlanPollClosed) {
		t.Fatalf("expected cancelled plan poll closed, got %v", err)
	}
}

func TestPlanMutationsPublishEventsToConversationMembers(t *testing.T) {
	channelID, creatorID, memberID := uuid.New(), uuid.New(), uuid.New()
	publisher := &planPublisherStub{}
	repo := &planRepoStub{}
	channelRepo := planChannelRepoStub{exists: true, member: true, members: []models.ChannelMember{{UserID: creatorID}, {UserID: memberID}}}
	userRepo := planUserRepoStub{records: []models.User{{ID: creatorID, Name: "Creator"}, {ID: memberID, Name: "Member"}}}
	service := NewPlanService(repo, channelRepo, planFriendRepoStub{}, userRepo, publisher)
	plan, err := service.CreateChannelPlan(context.Background(), creatorID, channelID, PlanInput{Title: "Meet", StartAt: time.Now().Add(48 * time.Hour)})
	if err != nil {
		t.Fatalf("CreateChannelPlan returned error: %v", err)
	}
	if len(publisher.published) != 1 || publisher.published[0].Event != events.PlanCreated || publisher.published[0].PlanID != plan.ID {
		t.Fatalf("missing plan-created event: %#v", publisher.published)
	}
	if len(publisher.published[0].Recipients) != 2 || !containsUUID(publisher.published[0].Recipients, creatorID) || !containsUUID(publisher.published[0].Recipients, memberID) {
		t.Fatalf("event recipients did not match conversation members: %#v", publisher.published[0].Recipients)
	}
	repo.plan = repo.created
	repo.plan.UpdatedAt = time.Now()
	if _, err := service.SetResponse(context.Background(), memberID, plan.ID, models.PlanResponseIn); err != nil {
		t.Fatalf("SetResponse returned error: %v", err)
	}
	if publisher.published[1].Event != events.PlanResponseChange {
		t.Fatalf("expected response-changed event, got %#v", publisher.published[1])
	}
	updated, err := service.UpdatePlan(context.Background(), creatorID, plan.ID, UpdatePlanInput{
		Title: "Meet updated", StartAt: plan.StartAt, ExpectedUpdatedAt: repo.plan.UpdatedAt,
	})
	if err != nil {
		t.Fatalf("UpdatePlan returned error: %v", err)
	}
	if publisher.published[2].Event != events.PlanUpdated || publisher.published[2].PlanID != updated.ID {
		t.Fatalf("expected plan-updated event, got %#v", publisher.published[2])
	}
	updated, err = service.UpdatePlanStatus(context.Background(), creatorID, plan.ID, models.PlanStatusConfirmed, updated.UpdatedAt)
	if err != nil {
		t.Fatalf("confirming plan returned error: %v", err)
	}
	if publisher.published[3].Event != events.PlanConfirmed {
		t.Fatalf("expected plan-confirmed event, got %#v", publisher.published[3])
	}
	_, err = service.UpdatePlanStatus(context.Background(), creatorID, plan.ID, models.PlanStatusCancelled, updated.UpdatedAt)
	if err != nil {
		t.Fatalf("cancelling plan returned error: %v", err)
	}
	if publisher.published[4].Event != events.PlanCancelled {
		t.Fatalf("expected plan-cancelled event, got %#v", publisher.published[4])
	}
}

func TestUnauthorizedPlanMutationDoesNotPublishEvent(t *testing.T) {
	planID, channelID, creatorID := uuid.New(), uuid.New(), uuid.New()
	publisher := &planPublisherStub{}
	repo := &planRepoStub{plan: &models.Plan{ID: planID, CreatorID: creatorID, ChannelID: &channelID, StartAt: time.Now().Add(48 * time.Hour), UpdatedAt: time.Now()}}
	service := NewPlanService(repo, planChannelRepoStub{exists: true, member: false}, planFriendRepoStub{}, planUserRepoStub{}, publisher)
	_, err := service.UpdatePlanStatus(context.Background(), creatorID, planID, models.PlanStatusConfirmed, repo.plan.UpdatedAt)
	if !errors.Is(err, ErrPlanAccessDenied) {
		t.Fatalf("expected denied plan update, got %v", err)
	}
	if len(publisher.published) != 0 {
		t.Fatalf("unauthorized update emitted events: %#v", publisher.published)
	}
}

func TestDMPlanCreatedEventTargetsBothConversationMembers(t *testing.T) {
	creatorID, friendID, outsiderID := uuid.New(), uuid.New(), uuid.New()
	publisher := &planPublisherStub{}
	service := NewPlanService(&planRepoStub{}, planChannelRepoStub{}, planFriendRepoStub{relation: &models.Friend{Status: StatusAccepted}}, planUserRepoStub{exists: true}, publisher)
	plan, err := service.CreateDMPlan(context.Background(), creatorID, friendID, PlanInput{Title: "Coffee", StartAt: time.Now().Add(48 * time.Hour)})
	if err != nil {
		t.Fatalf("CreateDMPlan returned error: %v", err)
	}
	if len(publisher.published) != 1 || publisher.published[0].Event != events.PlanCreated {
		t.Fatalf("missing DM plan-created event: %#v", publisher.published)
	}
	recipients := publisher.published[0].Recipients
	if len(recipients) != 2 || !containsUUID(recipients, creatorID) || !containsUUID(recipients, friendID) || containsUUID(recipients, outsiderID) {
		t.Fatalf("DM event recipients were not scoped to the conversation: %#v", recipients)
	}
	if publisher.published[0].DMUserLowID == nil || publisher.published[0].DMUserHighID == nil || plan.ID != publisher.published[0].PlanID {
		t.Fatalf("DM event did not identify the authorized conversation: %#v", publisher.published[0])
	}
}

func containsUUID(values []uuid.UUID, want uuid.UUID) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
