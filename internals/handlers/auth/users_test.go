package auth

import (
	"chat-server/internals/config"
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type searchUserRepository struct {
	repository.UserRepository
	query     string
	excluded  uuid.UUID
	limit     int
	users     []models.User
	searchErr error
}

func (r *searchUserRepository) SearchByUsername(_ context.Context, query string, excluded uuid.UUID, limit int) ([]models.User, error) {
	r.query, r.excluded, r.limit = query, excluded, limit
	return r.users, r.searchErr
}

func TestGetUsersSearchRequiresAuthenticationAndUsername(t *testing.T) {
	previousToken := config.Cfg.Secret.TOKEN
	config.Cfg.Secret.TOKEN = "test-token-secret"
	t.Cleanup(func() { config.Cfg.Secret.TOKEN = previousToken })

	userID := uuid.New()
	repo := &searchUserRepository{users: []models.User{{ID: uuid.New(), Name: "milo"}}}
	handler := http.Handler(middleware.AuthMiddleware(http.HandlerFunc(NewAuthHandler(repo).GetUsers)))

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/users?username=mi", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated request to return 401, got %d", unauthorized.Code)
	}

	badQuery := httptest.NewRecorder()
	token, err := utils.GenerateAccessToken(userID, "caller")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/users?username=m", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(badQuery, request)
	if badQuery.Code != http.StatusBadRequest {
		t.Fatalf("expected short query to return 400, got %d", badQuery.Code)
	}

	response := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/users?username=%20mi%20", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected valid search to return 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.query != "mi" || repo.excluded != userID || repo.limit != 20 {
		t.Fatalf("search was not constrained as expected: query=%q excluded=%s limit=%d", repo.query, repo.excluded, repo.limit)
	}
	var body struct {
		Users []map[string]any `json:"users"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Users) != 1 || body.Users[0]["username"] != "milo" {
		t.Fatalf("unexpected search response: %s", response.Body.String())
	}
	if _, includesEmail := body.Users[0]["email"]; includesEmail {
		t.Fatal("user search must not expose email addresses")
	}
}
