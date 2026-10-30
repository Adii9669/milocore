package messages

import (
	"chat-server/internals/middleware"
	"chat-server/internals/services"
	"chat-server/internals/websockets"
	"chat-server/storage"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service services.MessageService
	storage storage.FileStorage
	hub     *websockets.Hub
}

func NewHandler(service services.MessageService, fileStorage storage.FileStorage, hub *websockets.Hub) *Handler {
	return &Handler{service: service, storage: fileStorage, hub: hub}
}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	messageID, err := uuid.Parse(chi.URLParam(r, "messageID"))
	if err != nil {
		http.Error(w, "invalid message ID", http.StatusBadRequest)
		return
	}
	var input struct {
		Scope string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	event, objectKey, err := h.service.DeleteMessage(r.Context(), userID, messageID, strings.ToLower(strings.TrimSpace(input.Scope)))
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "not authorized") || strings.Contains(err.Error(), "only the sender") {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	if event.MessageID != uuid.Nil {
		if objectKey != nil && h.storage != nil {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.storage.DeleteObject(cleanupCtx, *objectKey); err != nil {
				log.Printf("failed to remove MinIO object for deleted message %s: %v", messageID, err)
			}
		}
		if h.hub != nil {
			h.hub.PublishMessageDeletion(event)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
