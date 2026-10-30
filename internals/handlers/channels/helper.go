package channels

import (
	"chat-server/internals/repository"
	"chat-server/internals/services"
	"errors"
	"net/http"
)

func (h *ChannelHandler) handleError(w http.ResponseWriter, err error) {

	switch {
	case errors.Is(err, services.ErrPermissionDenied):
		http.Error(w, err.Error(), http.StatusForbidden)

	case errors.Is(err, repository.ErrMemberNotFound),
		errors.Is(err, repository.ErrChannelNotFound),
		errors.Is(err, services.ErrMemberNotFound),
		errors.Is(err, services.ErrChannelNotFound),
		errors.Is(err, services.ErrUserNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	case errors.Is(err, services.ErrNameEmpty),
		errors.Is(err, services.ErrInvalidRole):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, services.ErrAlreadyMember):
		http.Error(w, err.Error(), http.StatusConflict)

	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
