package crews

import (
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func DeleteCrewHandler(repo repository.CrewRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 1. Check auth claims
		claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
		if !ok || claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Print("HI")

		// 2. Parse user ID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "Invalid User ID", http.StatusBadRequest)
			return
		}

		// 3. Extract crew ID using Gorilla Mux
		vars := mux.Vars(r)
		crewIDStr := vars["id"]
		crewID, err := uuid.Parse(crewIDStr)
		if err != nil {
			http.Error(w, "Invalid Crew ID", http.StatusBadRequest)
			return
		}

		// 4. Delete the crew
		err = repo.DeleteCrewByID(userID, crewID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Crew not found or not allowed", http.StatusForbidden)
				return
			}
			http.Error(w, "Failed to delete crew", http.StatusInternalServerError)
			return
		}

		// 5. Success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}
}
