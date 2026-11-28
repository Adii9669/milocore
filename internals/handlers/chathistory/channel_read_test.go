package chathistory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/middleware"
	"chat-server/internals/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type testReadStore struct{ user, channel, message uuid.UUID }

func (s *testReadStore) AdvanceChannelRead(_ context.Context, user, channel, message uuid.UUID) error {
	s.user, s.channel, s.message = user, channel, message
	return nil
}

func TestChannelReadMarkerUsesAuthenticatedMemberAndRoutedChannel(t *testing.T) {
	old := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "channel-read-test"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = old })

	user, channel, message := uuid.New(), uuid.New(), uuid.New()
	token, err := utils.GenerateAccessToken(user, "member")
	if err != nil {
		t.Fatal(err)
	}
	store := &testReadStore{}
	router := chi.NewRouter()
	router.Use(middleware.AuthMiddleware)
	router.Put("/api/chats/channel/{channelId}/read", UpdateChannelReadMarker(store))

	req := httptest.NewRequest(http.MethodPut, "/api/chats/channel/"+channel.String()+"/read",
		strings.NewReader(`{"messageId":"`+message.String()+`"}`))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent || store.user != user || store.channel != channel || store.message != message {
		t.Fatalf("incorrect channel read request: HTTP %d", w.Code)
	}
}
