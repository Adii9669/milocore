package chathistory

import (
	"chat-server/internals/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetCrewHistory(w http.ResponseWriter, r *http.Request) {

	crewID := chi.URLParam(r, "crewId")

	messages, err := h.service.CrewHistory(crewID, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, messages)
}
