package channels

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type channelRoleService struct {
	services.ChannelService
	channel, requester, member uuid.UUID
	role                       models.ChannelRole
}

func (s *channelRoleService) UpdateMemberRole(_ context.Context, channel, requester, member uuid.UUID, role models.ChannelRole) error {
	s.channel, s.requester, s.member, s.role = channel, requester, member, role
	return nil
}

func TestUpdateRoleUsesRoutedUserID(t *testing.T) {
	old := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "test-channel-role-token"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = old })

	channel, owner, member := uuid.New(), uuid.New(), uuid.New()
	token, err := utils.GenerateAccessToken(owner, "owner")
	if err != nil {
		t.Fatal(err)
	}
	service := &channelRoleService{}
	router := chi.NewRouter()
	router.Use(middleware.AuthMiddleware)
	router.Put("/api/channels/{channelID}/members/{userID}/role", NewChannelHandler(service, nil).UpdateMember)
	req := httptest.NewRequest(http.MethodPut, "/api/channels/"+channel.String()+"/members/"+member.String()+"/role",
		strings.NewReader(`{"role":"admin"}`))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)

	if response.Code != http.StatusOK || service.channel != channel || service.requester != owner ||
		service.member != member || service.role != models.RoleAdmin {
		t.Fatalf("unexpected role update: status %d, %s", response.Code, response.Body.String())
	}
}
