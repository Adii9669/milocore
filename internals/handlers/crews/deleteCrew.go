package crews

import (
	"chat-server/internals/middleware"
	"chat-server/internals/utils"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *CrewHandler) DeleteCrew(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	// 2. Parse user ID
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// 3. Extract crew ID using Gorilla Mux
	crewIDStr := chi.URLParam(r, "crewID")
	crewID, err := uuid.Parse(crewIDStr)
	if err != nil {
		http.Error(w, "Invalid Crew ID", http.StatusBadRequest)
		return
	}

	// 4. Delete the crew
	err = h.crewService.DeleteCrew(ctx, crewID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Crew not found", http.StatusForbidden)
			return
		}

		fmt.Printf("Error while deleting....", err)
		http.Error(w, "Failed to delete crew", http.StatusInternalServerError)
		return
	}

	// 5. Success
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]any{"success": true})
}
