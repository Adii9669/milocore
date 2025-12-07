package messages

import (
	"chat-server/internals/utils"
	"net/http"
)

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {

	//get the crew
	crewID := r.URL.Query().Get("crewId")
	if crewID == "" {
		http.Error(w, "crewID required", http.StatusBadRequest)
		return
	}

	messages, err := h.MessageRepo.GetMessagesByCrewID(crewID, 200)
	if err != nil {
		http.Error(w, "Failed to fetch messages.", http.StatusInternalServerError)
		return
	}

	utils.PrettyJSON(w, (messages))
}
