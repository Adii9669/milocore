package auth

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"net/http"
)

func (h *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	users, err := h.UserRepo.FindAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]any, 0, len(users))
	for _, u := range users {
		result = append(result, map[string]any{
			"id":    u.ID.String(),
			"name":  u.Name,
			"email": u.Email,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{
		"message": "success",
		"users":   result,
	})
}
