package plans

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePlanRequiresAuthentication(t *testing.T) {
	handler := NewHandler(nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/api/chats/channel/channel-id/plans", nil)
	handler.CreateChannelPlan(recorder, request)
	if recorder.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated request, got %d", recorder.Code)
	}
}

func TestListPlansRequiresAuthentication(t *testing.T) {
	handler := NewHandler(nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/chats/dm/user-id/plans", nil)
	handler.ListDMPlans(recorder, request)
	if recorder.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated request, got %d", recorder.Code)
	}
}

func TestParticipantHandlersRequireAuthentication(t *testing.T) {
	handler := NewHandler(nil)
	for _, test := range []struct {
		method string
		path   string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{method: "PUT", path: "/api/plans/plan-id/response", handle: handler.SetResponse},
		{method: "GET", path: "/api/plans/plan-id/participants", handle: handler.GetParticipants},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(test.method, test.path, nil)
		test.handle(recorder, request)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", test.method, test.path, recorder.Code)
		}
	}
}

func TestPlanDetailsAndUpdateHandlersRequireAuthentication(t *testing.T) {
	handler := NewHandler(nil)
	for _, test := range []struct {
		method string
		path   string
		handle func(http.ResponseWriter, *http.Request)
	}{
		{method: "GET", path: "/api/plans/id", handle: handler.GetPlan},
		{method: "GET", path: "/api/plans/id/poll", handle: handler.GetPoll},
		{method: "POST", path: "/api/plans/id/poll", handle: handler.CreatePoll},
		{method: "PUT", path: "/api/plans/id/poll/votes", handle: handler.SetPollVotes},
		{method: "PUT", path: "/api/plans/id", handle: handler.UpdatePlan},
		{method: "PUT", path: "/api/plans/id/status", handle: handler.UpdatePlanStatus},
	} {
		recorder := httptest.NewRecorder()
		test.handle(recorder, httptest.NewRequest(test.method, test.path, nil))
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", test.method, test.path, recorder.Code)
		}
	}
}

func TestDecodePlanPollRequestsRejectsUnknownOrMissingFields(t *testing.T) {
	for _, body := range []string{
		`{"question":"Pick","options":["A","B"],"planId":"other"}`,
		`{"question":"Pick","options":"A"}`,
	} {
		recorder := httptest.NewRecorder()
		if _, ok := decodeCreatePlanPollRequest(recorder, httptest.NewRequest("POST", "/", strings.NewReader(body))); ok || recorder.Code != http.StatusBadRequest {
			t.Errorf("expected invalid poll request to return 400, got ok=%v status=%d", ok, recorder.Code)
		}
	}
	for _, body := range []string{`{}`, `{"optionIds":null}`, `{"optionIds":[],"userId":"other"}`} {
		recorder := httptest.NewRecorder()
		if _, ok := decodePlanPollVotesRequest(recorder, httptest.NewRequest("PUT", "/", strings.NewReader(body))); ok || recorder.Code != http.StatusBadRequest {
			t.Errorf("expected invalid vote request to return 400, got ok=%v status=%d", ok, recorder.Code)
		}
	}
}

func TestDecodePlanStatusRequiresValidShapeAndRejectsClientAssociation(t *testing.T) {
	for _, body := range []string{
		`{"status":"confirmed"}`,
		`{"status":"scheduled","expectedUpdatedAt":"2030-01-02T14:00:00Z"}`,
		`{"status":"open","expectedUpdatedAt":"2030-01-02T14:00:00Z","creatorId":"other"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("PUT", "/api/plans/id/status", strings.NewReader(body))
		if _, ok := decodePlanStatusRequest(recorder, request); ok || recorder.Code != http.StatusBadRequest {
			t.Errorf("expected invalid status payload to return 400 (ok=%v status=%d)", ok, recorder.Code)
		}
	}
}

func TestDecodePlanUpdateRequiresConcurrencyTokenAndRejectsImmutableFields(t *testing.T) {
	for _, body := range []string{
		`{"title":"New title","startAt":"2030-01-02T15:04:05Z"}`,
		`{"title":"New title","startAt":"2030-01-02T15:04:05Z","expectedUpdatedAt":"2030-01-02T14:00:00Z","creatorId":"other"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("PUT", "/api/plans/id", strings.NewReader(body))
		if _, ok := decodePlanUpdateRequest(recorder, request); ok || recorder.Code != http.StatusBadRequest {
			t.Errorf("expected invalid update payload to return 400 (ok=%v status=%d)", ok, recorder.Code)
		}
	}
}

func TestDecodePlanResponseRequestValidatesStatusAndRejectsUnknownFields(t *testing.T) {
	for _, test := range []struct {
		body  string
		valid bool
	}{
		{body: `{"response":"in"}`, valid: true},
		{body: `{"response":"maybe"}`, valid: true},
		{body: `{"response":"cant_make_it"}`, valid: true},
		{body: `{"response":"no_response"}`},
		{body: `{"response":"reject"}`},
		{body: `{"response":"maybe","userId":"client-controlled"}`},
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("PUT", "/api/plans/id/response", strings.NewReader(test.body))
		_, ok := decodePlanResponseRequest(recorder, request)
		if ok != test.valid || (!test.valid && recorder.Code != http.StatusBadRequest) {
			t.Errorf("unexpected decode for %s (ok=%v status=%d)", test.body, ok, recorder.Code)
		}
	}
}

func TestDecodePlanRejectsInvalidTimestampAndUnknownField(t *testing.T) {
	for _, body := range []string{
		`{"title":"Plan","startAt":"not-a-time"}`,
		`{"title":"Plan","startAt":"2030-01-02T15:04:05Z","creatorId":"client-controlled"}`,
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/api/chats/channel/channel-id/plans", strings.NewReader(body))
		if _, ok := decodePlanRequest(recorder, request); ok || recorder.Code != 400 {
			t.Errorf("expected malformed request to return 400 (ok=%v status=%d)", ok, recorder.Code)
		}
	}
}
