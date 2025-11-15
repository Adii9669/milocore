package crews

import (
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"chat-server/middleware"
	"encoding/json"

	// "log"
	"net/http"
)

func Getcrew(repo repository.CrewRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//1. always get the claims first authentication from middleware
		claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
		if !ok || claims == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// log.Printf("DEBUG: Handling request for UserID: %s", claims.UserID)

		//2. check the user name from the details is that exist or not
		crews, err := repo.FindForUser(claims.UserID)
		// log.Printf("USEr BY username %v", crews)
		if err != nil {
			http.Error(w, "Failed to retrive the Crew", http.StatusUnauthorized)
			return
		}

		// 3. Handle empty result safely
		if len(crews) == 0 {
			// Always return an empty array instead of message
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			utils.PrettyJSON(w, []models.CrewResponse{})
			return
		}

		//3.Iterate through the response you got and send only the required response
		response := make([]models.CrewResponse, 0)
		for _, crew := range crews {
			response = append(response, models.CrewResponse{
				ID:        crew.ID,
				Name:      crew.Name,
				OwnerID:   crew.OwnerID,
				CreatedAt: crew.CreatedAt,
			})

		}

		//debugJSON
		// debugJSON, _ := json.MarshalIndent(response, "", "  ")
		// log.Printf("DEBUG: Sending response: \n%s", string(debugJSON))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	}
}
