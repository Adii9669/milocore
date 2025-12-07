package crews

import (
	"chat-server/internals/repository"
)

type CrewHandler struct {
	CrewRepo repository.CrewRepository
}

func NewCrewHandler(repo repository.CrewRepository) *CrewHandler {
	return &CrewHandler{
		CrewRepo: repo,
	}
}
