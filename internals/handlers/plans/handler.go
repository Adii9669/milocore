package plans

import (
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/transport/dto"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxPlanRequestSize = 1 << 20

type Handler struct{ service services.PlanService }

func NewHandler(service services.PlanService) *Handler { return &Handler{service: service} }

type planRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	StartAt     time.Time  `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	Location    *string    `json:"location"`
}

type planResponseRequest struct {
	Response models.PlanResponseStatus `json:"response"`
}

type planUpdateRequest struct {
	Title             string     `json:"title"`
	Description       *string    `json:"description"`
	StartAt           time.Time  `json:"startAt"`
	EndAt             *time.Time `json:"endAt"`
	Location          *string    `json:"location"`
	ExpectedUpdatedAt time.Time  `json:"expectedUpdatedAt"`
}

type planStatusRequest struct {
	Status            string    `json:"status"`
	ExpectedUpdatedAt time.Time `json:"expectedUpdatedAt"`
}

type createPlanPollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type planPollVotesRequest struct {
	OptionIDs []string `json:"optionIds"`
}

func (h *Handler) GetPoll(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	poll, err := h.service.GetPoll(r.Context(), userID, planID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toPlanPollResponse(poll))
}

func (h *Handler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	input, ok := decodeCreatePlanPollRequest(w, r)
	if !ok {
		return
	}
	poll, err := h.service.CreatePoll(r.Context(), userID, planID, input.Question, input.Options)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusCreated, toPlanPollResponse(poll))
}

func (h *Handler) SetPollVotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanPollVotesRequest(w, r)
	if !ok {
		return
	}
	optionIDs := make([]uuid.UUID, 0, len(input.OptionIDs))
	for _, value := range input.OptionIDs {
		optionID, err := uuid.Parse(value)
		if err != nil {
			http.Error(w, "optionIds must contain valid IDs", http.StatusBadRequest)
			return
		}
		optionIDs = append(optionIDs, optionID)
	}
	poll, err := h.service.SetPollVotes(r.Context(), userID, planID, optionIDs)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toPlanPollResponse(poll))
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	plan, err := h.service.GetPlan(r.Context(), userID, planID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toPlanResponse(plan))
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanUpdateRequest(w, r)
	if !ok {
		return
	}
	plan, err := h.service.UpdatePlan(r.Context(), userID, planID, services.UpdatePlanInput{
		Title: input.Title, Description: input.Description, StartAt: input.StartAt,
		EndAt: input.EndAt, Location: input.Location, ExpectedUpdatedAt: input.ExpectedUpdatedAt,
	})
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toPlanResponse(plan))
}

func (h *Handler) UpdatePlanStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanStatusRequest(w, r)
	if !ok {
		return
	}
	plan, err := h.service.UpdatePlanStatus(r.Context(), userID, planID, input.Status, input.ExpectedUpdatedAt)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toPlanResponse(plan))
}

func (h *Handler) SetResponse(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanResponseRequest(w, r)
	if !ok {
		return
	}
	summary, err := h.service.SetResponse(r.Context(), userID, planID, input.Response)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toParticipationResponse(summary))
}

func (h *Handler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		http.Error(w, "invalid plan ID", http.StatusBadRequest)
		return
	}
	summary, err := h.service.GetParticipants(r.Context(), userID, planID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusOK, toParticipationResponse(summary))
}

func (h *Handler) CreateChannelPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		http.Error(w, "invalid channel ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanRequest(w, r)
	if !ok {
		return
	}
	plan, err := h.service.CreateChannelPlan(r.Context(), userID, channelID, toPlanInput(input))
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusCreated, toPlanResponse(plan))
}

func (h *Handler) ListChannelPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	channelID, err := uuid.Parse(chi.URLParam(r, "channelId"))
	if err != nil {
		http.Error(w, "invalid channel ID", http.StatusBadRequest)
		return
	}
	plans, err := h.service.ListChannelPlans(r.Context(), userID, channelID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	response := make([]dto.PlanResponse, 0, len(plans))
	for i := range plans {
		response = append(response, toPlanResponse(&plans[i]))
	}
	writePlanJSON(w, http.StatusOK, response)
}

func (h *Handler) CreateDMPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	otherUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	input, ok := decodePlanRequest(w, r)
	if !ok {
		return
	}
	plan, err := h.service.CreateDMPlan(r.Context(), userID, otherUserID, toPlanInput(input))
	if err != nil {
		writePlanError(w, err)
		return
	}
	writePlanJSON(w, http.StatusCreated, toPlanResponse(plan))
}

func (h *Handler) ListDMPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUser(w, r)
	if !ok {
		return
	}
	otherUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	plans, err := h.service.ListDMPlans(r.Context(), userID, otherUserID)
	if err != nil {
		writePlanError(w, err)
		return
	}
	response := make([]dto.PlanResponse, 0, len(plans))
	for i := range plans {
		response = append(response, toPlanResponse(&plans[i]))
	}
	writePlanJSON(w, http.StatusOK, response)
}

func authenticatedUser(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return uuid.Nil, false
	}
	return userID, true
}

func decodePlanRequest(w http.ResponseWriter, r *http.Request) (planRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input planRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid plan request: "+err.Error(), http.StatusBadRequest)
		return planRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return planRequest{}, false
	}
	return input, true
}

func decodePlanResponseRequest(w http.ResponseWriter, r *http.Request) (planResponseRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input planResponseRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid response request: "+err.Error(), http.StatusBadRequest)
		return planResponseRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return planResponseRequest{}, false
	}
	if !input.Response.IsValid() {
		http.Error(w, "response must be in, maybe, or cant_make_it", http.StatusBadRequest)
		return planResponseRequest{}, false
	}
	return input, true
}

func decodePlanUpdateRequest(w http.ResponseWriter, r *http.Request) (planUpdateRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input planUpdateRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid plan update request: "+err.Error(), http.StatusBadRequest)
		return planUpdateRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return planUpdateRequest{}, false
	}
	if input.ExpectedUpdatedAt.IsZero() {
		http.Error(w, "expectedUpdatedAt is required", http.StatusBadRequest)
		return planUpdateRequest{}, false
	}
	return input, true
}

func decodePlanStatusRequest(w http.ResponseWriter, r *http.Request) (planStatusRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input planStatusRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid plan status request: "+err.Error(), http.StatusBadRequest)
		return planStatusRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return planStatusRequest{}, false
	}
	if !models.IsValidPlanStatus(input.Status) {
		http.Error(w, "status must be open, confirmed, completed, or cancelled", http.StatusBadRequest)
		return planStatusRequest{}, false
	}
	if input.ExpectedUpdatedAt.IsZero() {
		http.Error(w, "expectedUpdatedAt is required", http.StatusBadRequest)
		return planStatusRequest{}, false
	}
	return input, true
}

func decodeCreatePlanPollRequest(w http.ResponseWriter, r *http.Request) (createPlanPollRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input createPlanPollRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid poll request: "+err.Error(), http.StatusBadRequest)
		return createPlanPollRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return createPlanPollRequest{}, false
	}
	return input, true
}

func decodePlanPollVotesRequest(w http.ResponseWriter, r *http.Request) (planPollVotesRequest, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanRequestSize)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input planPollVotesRequest
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid poll vote request: "+err.Error(), http.StatusBadRequest)
		return planPollVotesRequest{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain one JSON object", http.StatusBadRequest)
		return planPollVotesRequest{}, false
	}
	if input.OptionIDs == nil {
		http.Error(w, "optionIds is required", http.StatusBadRequest)
		return planPollVotesRequest{}, false
	}
	return input, true
}

func toPlanInput(input planRequest) services.PlanInput {
	return services.PlanInput{Title: input.Title, Description: input.Description, StartAt: input.StartAt, EndAt: input.EndAt, Location: input.Location}
}

func toPlanResponse(plan *models.Plan) dto.PlanResponse {
	response := dto.PlanResponse{ID: plan.ID.String(), Title: plan.Title, Description: plan.Description, CreatorID: plan.CreatorID.String(), CreatorName: plan.Creator.Name, StartAt: plan.StartAt, EndAt: plan.EndAt, Location: plan.Location, Status: plan.Status, CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt}
	if plan.ChannelID != nil {
		value := plan.ChannelID.String()
		response.ChannelID = &value
	}
	if plan.DMUserLowID != nil {
		value := plan.DMUserLowID.String()
		response.DMUserLowID = &value
	}
	if plan.DMUserHighID != nil {
		value := plan.DMUserHighID.String()
		response.DMUserHighID = &value
	}
	return response
}

func toParticipationResponse(summary *services.PlanParticipationSummary) dto.PlanParticipationSummary {
	response := dto.PlanParticipationSummary{
		PlanID: summary.PlanID.String(), InCount: summary.InCount, MaybeCount: summary.MaybeCount,
		CantMakeItCount: summary.CantMakeItCount, PendingResponseCount: summary.PendingResponseCount,
		Participants: make([]dto.PlanParticipantResponse, 0, len(summary.Participants)),
	}
	for _, participant := range summary.Participants {
		response.Participants = append(response.Participants, dto.PlanParticipantResponse{
			UserID: participant.UserID.String(), Name: participant.Name, Image: participant.Image,
			Response: participant.Response, RespondedAt: participant.RespondedAt,
		})
	}
	return response
}

func toPlanPollResponse(poll *services.PlanPollSummary) dto.PlanPollResponse {
	response := dto.PlanPollResponse{ID: poll.ID.String(), PlanID: poll.PlanID.String(), Question: poll.Question, CreatedAt: poll.CreatedAt, UpdatedAt: poll.UpdatedAt, Options: make([]dto.PlanPollOptionResponse, 0, len(poll.Options))}
	for _, option := range poll.Options {
		response.Options = append(response.Options, dto.PlanPollOptionResponse{ID: option.ID.String(), Text: option.Text, VoteCount: option.VoteCount, SelectedByMe: option.SelectedByMe})
	}
	return response
}

func writePlanJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writePlanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidPlan):
		http.Error(w, "title and a valid startAt are required; endAt must be later than startAt", http.StatusBadRequest)
	case errors.Is(err, services.ErrPlanChatNotFound):
		http.Error(w, "chat not found", http.StatusNotFound)
	case errors.Is(err, services.ErrPlanNotFound):
		http.Error(w, "plan not found", http.StatusNotFound)
	case errors.Is(err, services.ErrPlanAccessDenied):
		http.Error(w, "not authorized to access this chat", http.StatusForbidden)
	case errors.Is(err, services.ErrPlanCreatorOnly):
		http.Error(w, "only the plan creator can manage this plan", http.StatusForbidden)
	case errors.Is(err, services.ErrPlanStale):
		http.Error(w, "plan has changed; reload it before updating", http.StatusConflict)
	case errors.Is(err, services.ErrPlanCancelled):
		http.Error(w, "cancelled plans cannot be changed", http.StatusConflict)
	case errors.Is(err, services.ErrInvalidPlanStatus):
		http.Error(w, "status must be open, confirmed, completed, or cancelled", http.StatusBadRequest)
	case errors.Is(err, services.ErrInvalidPlanTransition):
		http.Error(w, "invalid plan status transition", http.StatusConflict)
	case errors.Is(err, services.ErrInvalidPlanResponse):
		http.Error(w, "response must be in, maybe, or cant_make_it", http.StatusBadRequest)
	case errors.Is(err, services.ErrPlanNotRespondable):
		http.Error(w, "plan is not accepting responses", http.StatusConflict)
	case errors.Is(err, services.ErrInvalidPlanPoll):
		http.Error(w, "poll requires a question and 2 to 10 distinct nonempty options", http.StatusBadRequest)
	case errors.Is(err, services.ErrPlanPollExists):
		http.Error(w, "this plan already has a poll", http.StatusConflict)
	case errors.Is(err, services.ErrPlanPollNotFound):
		http.Error(w, "poll not found", http.StatusNotFound)
	case errors.Is(err, services.ErrPlanPollClosed):
		http.Error(w, "poll is not accepting votes", http.StatusConflict)
	default:
		http.Error(w, "failed to process plan", http.StatusInternalServerError)
	}
}
