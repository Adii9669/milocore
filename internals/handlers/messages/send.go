package messages

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/transport/dto"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input services.IncomingMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, "invalid message request", http.StatusBadRequest)
		return
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid message request", http.StatusBadRequest)
		return
	}

	input.Content = strings.TrimSpace(input.Content)
	if input.Content == "" || len(input.Content) > 4<<10 {
		http.Error(w, "message must contain between 1 and 4096 bytes", http.StatusBadRequest)
		return
	}

	event, err := h.service.HandleIncomingMessage(r.Context(), userID, input)
	if err != nil {
		http.Error(w, "Onechat could not send this message", http.StatusBadRequest)
		return
	}
	if h.hub != nil {
		h.hub.PublishMessage(event)
	}

	isMine := true
	message := dto.MessageResponse{
		ID: event.ID.String(), Type: input.Type, Content: event.Content,
		CreatedAt: event.CreatedAt, IsMine: &isMine,
		Sender: &dto.SenderDTO{ID: event.SenderID.String(), Name: event.SenderName},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(message)
}
