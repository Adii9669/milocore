package crews

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"

	// "log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CrewHandler) Getcrew(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	//1. always get the claims first authentication from middleware
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
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

func (h *CrewHandler) GetMembers(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	//1. always get the claims first authentication from middleware
	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	crewIDStr := chi.URLParam(r, "crewID")
	crewID, err := uuid.Parse(crewIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// var req requests.GetMembers
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
	// 	return
	// }

	response, err := h.crewService.GetMembers(ctx, crewID, userID)
	if err != nil {
		http.Error(w, "Failed to retrive the Crew", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	utils.PrettyJSON(w, map[string]any{
		"members": response,
	})
}
