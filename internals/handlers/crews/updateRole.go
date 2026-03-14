package crews

import (
	"chat-server/internals/db/models"
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

func (h *CrewHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {

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

	//3.Parse memberID from URL
	memberIDParam := chi.URLParam(r, "memberID")
	memberID, err := uuid.Parse(memberIDParam)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// 4.check body for more info
	var req requests.CrewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Role) == "" {
		http.Error(w, "role is required", http.StatusBadRequest)
		return
	}

	newRole := models.CrewRole(req.Role)

	//call the service
	err = h.crewService.UpdateMemberRole(ctx, crewID, userID, memberID, newRole)
	if err != nil {
		log.Println("AddMember error:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{
		"success": "true",
	})
}
