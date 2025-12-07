package requests

type CreateCrewRequest struct {
	Name string `json:"name" binding:"required"`
}
