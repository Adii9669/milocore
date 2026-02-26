package crews

import (
	"chat-server/internals/middleware"
	"chat-server/internals/requests"
	"chat-server/internals/utils"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *CrewHandler) RenameCrew(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	ctx := r.Context()

	//1.Get userID
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

	// 4.check body for more info
	var req requests.CrewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	newName := req.Name

	//call the service
	err = h.crewService.RenameCrew(ctx, crewID, userID, newName)
	if err != nil {
		log.Println("Renaming error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{
		"message": "renamed Successfully",
	})
}
