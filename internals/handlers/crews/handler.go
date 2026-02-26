package crews

import (
	"chat-server/internals/services"
)

type CrewHandler struct {
	crewService services.CrewService
}

func NewCrewHandler(crewService services.CrewService) *CrewHandler {
	return &CrewHandler{
		crewService: crewService,
	}
}
