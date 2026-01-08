package requests

type CreateConvoRequest struct {
	UserTwoID string `json:"userTwoID" validate:"required,uuid4"`
}

