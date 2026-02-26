package crews

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CrewHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	// 1.Get the Owner ID
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusUnauthorized)
		return
	}

	//2.Parse crewID from URL
	crewIDParam := chi.URLParam(r, "crewID")
	crewID, err := uuid.Parse(crewIDParam)
	if err != nil {
		http.Error(w, "Invalid crew ID", http.StatusBadRequest)
		return
	}

	//3.Parse memberID from URL
	memberIDParam := chi.URLParam(r, "memberID")
	memberID, err := uuid.Parse(memberIDParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// // 4.check body for more info
	// var request requests.UpdateCrewMember
	// if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
	// 	http.Error(w, "Invalid user ID", http.StatusBadRequest)
	// 	return
	// }

	//call the service
	err = h.crewService.RemoveMember(ctx, crewID, userID, memberID)
	if err != nil {
		log.Println("AddMember error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{
		"message": "removed sucessfully",
	})
}
