package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	//internals
	"chat-server/internals/services"
	"chat-server/internals/utils"

	//libraries
	"github.com/go-playground/validator/v10"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required"`
}

var validate = validator.New()

func RegisterHandler(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//1.Decoding the Request Body
		var req RegisterRequest
		ctx := r.Context()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid body ", http.StatusBadRequest)
			return
		}

		//2.check the email or string is not empty
		//TrimSpace of the info (username , email ) recieved
		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(req.Email)

		//3.validation the request
		if err := validate.Struct(req); err != nil {
			//extract validation details to send
			validationError := err.(validator.ValidationErrors)
			errorsMap := make(map[string]string)

			for _, feildErr := range validationError {
				errorsMap[feildErr.Field()] = feildErr.Error()
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"validationError": errorsMap,
			})
			return
		}

		err := authService.Register(ctx, req.Username, req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		//success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		utils.PrettyJSON(w, map[string]any{
			"message": "Registration successful. Please check your email to verify your account.",
		})
	}

}
