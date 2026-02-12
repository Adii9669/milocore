package chathistory

import (
	"chat-server/internals/utils"
	"chat-server/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) GetCrewHistory(w http.ResponseWriter, r *http.Request) {

	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok {
		http.Error(w, "Not a valid User", http.StatusInternalServerError)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	crewID := chi.URLParam(r, "crewId")

	messages, err := h.service.CrewHistory(crewID, userID.String(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, messages)
}
