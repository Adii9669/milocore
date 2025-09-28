package crews

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func CreateCrewHandler(repo repository.CrewRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//1.get the authentication id
		claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
		if !ok || claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// ADD THIS LOG STATEMENT
		log.Printf("DEBUG: GetCrewsHandler received UserID from token: '%s'", claims.UserID)
		//2. Decode the body of the request
		var req CreateCrewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Body Request", http.StatusBadRequest)
			return
		}

		//parsing the jwtclaims userid (stored in the stirng type) to uuid format
		ownerID, err := uuid.Parse(claims.UserID)
		if err != nil {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		log.Printf("THE USERID %v", ownerID)

		//3.Create New Crew
		newCrew := &models.Crew{
			Name:    req.Name,
			OwnerID: ownerID,
		}

		//call the repo to save it
		if err := repo.Create(newCrew); err != nil {
			http.Error(w, "Failed to create the Crew", http.StatusInternalServerError)
			return
		}

		response := CrewResponse{
			ID:        newCrew.ID,
			Name:      newCrew.Name,
			OwnerID:   newCrew.OwnerID,
			CreatedAt: newCrew.CreatedAt,
		}

		//send sucess
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
