package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"chat-server/internals/config"
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/repository"
	"chat-server/internals/services"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type contractUserRepo struct {
	repository.UserRepository
	user models.User
}

func (r contractUserRepo) FindBYName(_ context.Context, _ string) (*models.User, error) {
	return &r.user, nil
}

func (r contractUserRepo) FindByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return &r.user, nil
}

type contractSessionRepo struct {
	repository.SessionRepository
	sessions  map[uuid.UUID]*models.Session
	lookupErr error
}

func (r *contractSessionRepo) Create(_ context.Context, session *models.Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *contractSessionRepo) FindByTokenHash(_ context.Context, hash string) (*models.Session, error) {
	if r.lookupErr != nil {
		return nil, r.lookupErr
	}
	for _, session := range r.sessions {
		if session.RefreshTokenHash == hash {
			return session, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *contractSessionRepo) DeleteSession(_ context.Context, userID, id uuid.UUID) error {
	if session := r.sessions[id]; session != nil && session.UserID == userID {
		delete(r.sessions, id)
	}
	return nil
}

func TestCookieSessionContract(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	previousToken := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "session-contract-test-secret"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = previousToken })

	hash, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	passwordHash := string(hash)
	users := contractUserRepo{user: models.User{ID: uuid.New(), Name: "milo", PasswordHash: &passwordHash}}
	sessions := &contractSessionRepo{sessions: make(map[uuid.UUID]*models.Session)}
	service := services.NewAuthService(users, sessions, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", LoginHandler(service))
	mux.HandleFunc("/auth/refresh", RefreshHandler(service))
	mux.Handle("/api/me", middleware.AuthMiddleware(http.HandlerFunc(NewAuthHandler(users).Me)))
	mux.Handle("/api/logout", middleware.AuthMiddleware(LogoutHandler(service)))
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	requestWith := func(jar http.CookieJar, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "https://onechat.test"+path, strings.NewReader(body))
		for _, cookie := range jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		jar.SetCookies(req.URL, response.Result().Cookies())
		return response
	}
	request := func(method, path, body string) *httptest.ResponseRecorder {
		return requestWith(jar, method, path, body)
	}
	requireStatus := func(response *httptest.ResponseRecorder, expected int) {
		t.Helper()
		if response.Code != expected {
			t.Fatalf("status = %d, want %d. Body: %s", response.Code, expected, response.Body)
		}
	}

	requireStatus(request(http.MethodPost, "/auth/login", `{"username":"milo","password":"test-password"}`), http.StatusOK)
	requireStatus(request(http.MethodGet, "/api/me", ""), http.StatusOK)

	apiRequest := httptest.NewRequest(http.MethodGet, "https://onechat.test/api/me", nil)
	cookies := jar.Cookies(apiRequest.URL)
	if len(cookies) != 1 || cookies[0].Name != "access_token" {
		t.Fatal("API requests should carry only the access cookie")
	}
	jar.SetCookies(apiRequest.URL, []*http.Cookie{{Name: "access_token", Path: "/", MaxAge: -1}})

	response := request(http.MethodGet, "/api/me", "")
	requireStatus(response, http.StatusUnauthorized)
	if response.Header().Get("X-Token-Expired") != "true" {
		t.Fatal("missing access cookie should trigger refresh")
	}
	sessions.lookupErr = errors.New("temporary database failure")
	requireStatus(request(http.MethodPost, "/auth/refresh", ""), http.StatusServiceUnavailable)
	sessions.lookupErr = nil
	requireStatus(request(http.MethodPost, "/auth/refresh", ""), http.StatusOK)
	requireStatus(request(http.MethodGet, "/api/me", ""), http.StatusOK)

	refreshRequest := httptest.NewRequest(http.MethodPost, "https://onechat.test/auth/refresh", nil)
	oldCookies := jar.Cookies(refreshRequest.URL)
	other, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	requireStatus(requestWith(other, http.MethodPost, "/auth/login", `{"username":"milo","password":"test-password"}`), http.StatusOK)
	requireStatus(request(http.MethodPost, "/auth/refresh", ""), http.StatusOK)
	requireStatus(request(http.MethodPost, "/api/logout", ""), http.StatusOK)
	requireStatus(request(http.MethodGet, "/api/me", ""), http.StatusUnauthorized)
	for _, cookie := range oldCookies {
		refreshRequest.AddCookie(cookie)
	}
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, refreshRequest)
	requireStatus(response, http.StatusUnauthorized)
	requireStatus(requestWith(other, http.MethodPost, "/auth/refresh", ""), http.StatusOK)
	requireStatus(requestWith(other, http.MethodGet, "/api/me", ""), http.StatusOK)
}
