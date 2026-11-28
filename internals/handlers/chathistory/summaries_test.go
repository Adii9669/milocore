package chathistory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/middleware"
	"chat-server/internals/utils"

	"github.com/google/uuid"
)

type testSummaryStore struct {
	account uuid.UUID
	rows    []ConversationSummary
}

func (s *testSummaryStore) List(_ context.Context, account uuid.UUID) ([]ConversationSummary, error) {
	s.account = account
	return s.rows, nil
}

func TestConversationSummariesUseAuthenticatedAccount(t *testing.T) {
	old := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "conversation-summary-test"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = old })

	account := uuid.New()
	token, err := utils.GenerateAccessToken(account, "caller")
	if err != nil {
		t.Fatal(err)
	}
	unread := int64(2)
	store := &testSummaryStore{rows: []ConversationSummary{
		{ID: uuid.New(), Kind: "dm", Name: "Friend", Unread: &unread},
		{ID: uuid.New(), Kind: "channel", Name: "Channel"},
	}}
	handler := middleware.AuthMiddleware(ListSummaries(store))
	req := httptest.NewRequest(http.MethodGet, "/api/conversations?userId="+uuid.NewString(), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != http.StatusUnauthorized || store.account != uuid.Nil {
		t.Fatal("unauthenticated request reached the summary store")
	}

	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	var rows []ConversationSummary
	if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || store.account != account || len(rows) != 2 ||
		rows[0].Unread == nil || *rows[0].Unread != 2 || rows[1].Unread != nil {
		t.Fatalf("unexpected conversation summaries: %s", response.Body.String())
	}
}
