package messages

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/events"
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/transport/dto"
	"chat-server/internals/utils"

	"github.com/google/uuid"
)

type sendService struct {
	services.MessageService
	calls   int
	sender  uuid.UUID
	message services.IncomingMessage
	event   events.MessageEvent
}

func (s *sendService) HandleIncomingMessage(_ context.Context, sender uuid.UUID, message services.IncomingMessage) (events.MessageEvent, error) {
	s.calls++
	s.sender = sender
	s.message = message
	s.event.SenderID = sender
	s.event.Content = message.Content
	return s.event, nil
}

func TestSendReturnsPersistedMessage(t *testing.T) {
	previousToken := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "send-contract-test-secret"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = previousToken })

	sender, recipient := uuid.New(), uuid.New()
	token, err := utils.GenerateAccessToken(sender, "sender")
	if err != nil {
		t.Fatal(err)
	}
	service := &sendService{event: events.MessageEvent{ID: uuid.New(), SenderName: "sender"}}
	handler := middleware.AuthMiddleware(http.HandlerFunc(NewHandler(service, nil, nil).Send))

	request := func(body string, authenticated bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(body))
		if authenticated {
			req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}

	body := `{"type":"dm","receiverId":"` + recipient.String() + `","content":" hello "}`
	if got := request(body, false).Code; got != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", got)
	}
	if got := request(body+`{}`, true).Code; got != http.StatusBadRequest {
		t.Fatalf("malformed status = %d", got)
	}
	if service.calls != 0 {
		t.Fatal("invalid requests reached the message service")
	}

	response := request(body, true)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body)
	}
	var saved dto.MessageResponse
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if service.calls != 1 || service.sender != sender || service.message.ReceiverID == nil ||
		*service.message.ReceiverID != recipient || saved.ID != service.event.ID.String() ||
		saved.Content != "hello" || saved.IsMine == nil || !*saved.IsMine {
		t.Fatalf("unexpected send result: %+v", saved)
	}
}
