package emails

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// internals
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/utils"

	//libraries
	// "github.com/gohugoio/hugo/config"
	"gorm.io/gorm"
)

type ResendRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	var req ResendRequest
	config := config.Cfg

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	//check the user
	var user db.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {

		//this for security if the unauthorized get's the emails but data does not exist still we show sucess message
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "If an account with that email exists, a new verification link has been sent.",
			})
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	//3.Check if the user is already verified
	if user.Verified {
		http.Error(w, "This account is already verified", http.StatusBadRequest)
	}

	//4. Generate the new token and send the email
	newVerificationToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		http.Error(w, "Failed to Generate The Token", http.StatusInternalServerError)
		return
	}

	//5.Update the user with new token
	user.VerifyToken = &newVerificationToken
	if err := db.DB.Save(&user).Error; err != nil {
		http.Error(w, "Failed to update the record", http.StatusInternalServerError)
		return
	}

	//6.build the base urls and send the mail
	baseURL := config.Server.APIBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	verificationURL := fmt.Sprintf("%s/api/auth/verify?token=%s", baseURL, newVerificationToken)
	if err := utils.SendVerificationEmail(*user.Email, verificationURL); err != nil {
		log.Printf("Failed to resend verification email for user %s: %v", user.ID, err)
		http.Error(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}
	// 7. Confirm to the user that the email has been sent.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "A new verification link has been sent to your email address.",
	})

}
