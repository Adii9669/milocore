package crews

import (
	"chat-server/internals/utils"
	"chat-server/middleware"

	// "log"
	"net/http"

	"github.com/google/uuid"
)

func (h *CrewHandler) Getcrew(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	//1. always get the claims first authentication from middleware
	ctx := r.Context()
	claims, ok := ctx.Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok || claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	response, err := h.crewService.GetCrews(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to retrive the Crew", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.PrettyJSON(w, map[string]any{
		"message": "Crews fetched successfully",
		"crews":   response,
	})
}
