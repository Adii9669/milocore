package push

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/utils"

	"github.com/google/uuid"
)

type recordingDeviceStore struct {
	device      models.PushDevice
	removedUser uuid.UUID
	removedID   uuid.UUID
}

func (s *recordingDeviceStore) register(_ context.Context, device models.PushDevice) error {
	s.device = device
	return nil
}
func (s *recordingDeviceStore) remove(_ context.Context, userID, id uuid.UUID) error {
	s.removedUser, s.removedID = userID, id
	return nil
}

func TestPushRegistrationAndRemovalUseAuthenticatedAccount(t *testing.T) {
	old := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "push-registration-test"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = old })

	account, installation := uuid.New(), uuid.New()
	token, err := utils.GenerateAccessToken(account, "owner")
	if err != nil {
		t.Fatal(err)
	}
	store := &recordingDeviceStore{}
	handler := Handler{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/devices/push-token", handler.Register)
	mux.HandleFunc("DELETE /api/devices/push-token", handler.Remove)
	protected := middleware.AuthMiddleware(mux)

	request := func(method, body string, authenticated bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/api/devices/push-token", strings.NewReader(body))
		if authenticated {
			r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		}
		w := httptest.NewRecorder()
		protected.ServeHTTP(w, r)
		return w
	}

	base := `{"installationId":"` + installation.String() + `"`
	input := base + `,"token":"test-fcm-token","platform":"android"}`
	if w := request("POST", input, false); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated registration: %d", w.Code)
	}
	if w := request("POST", base+`,"token":"abc","platform":"android","userId":"other"}`, true); w.Code != http.StatusBadRequest {
		t.Fatalf("client-supplied account was accepted: %d", w.Code)
	}
	if w := request("POST", input, true); w.Code != http.StatusNoContent {
		t.Fatalf("register: %d [%s]", w.Code, w.Body)
	}
	if store.device.UserID != account || store.device.InstallationID != installation {
		t.Fatal("registration was not scoped to the authenticated account and installation")
	}
	if w := request("DELETE", base+`}`, true); w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d [%s]", w.Code, w.Body)
	}
	if store.removedUser != account || store.removedID != installation {
		t.Fatal("removal was not scoped to the authenticated account and installation")
	}
}
