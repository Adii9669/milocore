package friends

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/utils"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *FriendHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusUnauthorized)
		return
	}

	//2.Take user id from the param
	friendIDStr := chi.URLParam(r, "requestID")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	err = h.friendService.RejectRequest(ctx, userID, friendID)
	if err != nil {
		log.Println("Reject Request Failed", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, map[string]string{
		"success": "true",
	})
}

func (h *FriendHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	sender, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	recipient, err := uuid.Parse(chi.URLParam(r, "requestID"))
	if err != nil || recipient == uuid.Nil || recipient == sender {
		http.Error(w, "invalid request ID", http.StatusBadRequest)
		return
	}
	err = h.friendService.CancelRequest(r.Context(), sender, recipient)
	if errors.Is(err, services.ErrRequestNotFound) {
		http.Error(w, "outgoing pending request not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not cancel the request", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
