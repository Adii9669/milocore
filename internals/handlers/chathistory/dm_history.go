package chathistory

import (
	"chat-server/internals/utils"
	"chat-server/middleware"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetDmHistory(w http.ResponseWriter, r *http.Request) {

	otherUserID := chi.URLParam(r, "userId")

	claims, ok := r.Context().Value(middleware.UserContextKey).(*utils.JWTClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	authUserID := claims.UserID

	log.Printf("userA %s", otherUserID)
	log.Printf("userB %s", authUserID)

	messages, err := h.service.DmHistory(authUserID, otherUserID, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	utils.PrettyJSON(w, messages)

}
