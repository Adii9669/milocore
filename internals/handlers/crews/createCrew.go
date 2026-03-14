package crews

import (
	"chat-server/internals/middleware"
	"chat-server/internals/requests"
	"chat-server/internals/utils"
	"encoding/json"
	"log"
	"net/http"
)

func (h *CrewHandler) CreateCrew(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	//1.get the authentication id
	ctx := r.Context()

	//2. Decode the body of the request
	var req requests.CrewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Body Request", http.StatusBadRequest)
		return
	}

	//parsing the jwtclaims userid (stored in the stirng type) to uuid format
	ownerID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//call the service
	response, err := h.crewService.CreateCrew(ctx, ownerID, req.Name)
	if err != nil {
		log.Println("CreateCrew error:", err)
		http.Error(w, "Failed To Create Crew", http.StatusInternalServerError)
		return
	}

	//send sucess
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	utils.PrettyJSON(w, response)
}
