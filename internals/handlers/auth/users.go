package auth

import (
	"chat-server/internals/db/models"
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("username"))
	if len([]rune(query)) < 2 || len([]rune(query)) > 40 {
		http.Error(w, "username search must be between 2 and 40 characters", http.StatusBadRequest)
		return
	}
	searchRepo, ok := h.UserRepo.(interface {
		SearchByUsername(context.Context, string, uuid.UUID, int) ([]models.User, error)
	})
	if !ok {
		http.Error(w, "Username search is unavailable", http.StatusInternalServerError)
		return
	}
	users, err := searchRepo.SearchByUsername(ctx, query, userID, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]any{"id": u.ID.String(), "username": u.Name})
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{
		"message": "success",
		"users":   result,
	})
}
