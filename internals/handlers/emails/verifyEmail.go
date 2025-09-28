package emails

import (
	"net/http"

	//internals
	"chat-server/internals/db"
	"chat-server/internals/db/models"
	"chat-server/internals/utils"

	//libraries
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {

	//get the token
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification Token Missing", http.StatusBadRequest)
		return
	}

	//find the user in the db
	var user models.User
	result := db.DB.Where("verify_token = ? ", token).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid or expired varification token.", http.StatusNotFound)
			return
		}
		http.Error(w, "DataBase Error", http.StatusInternalServerError)
		return
	}

	//update the user to verified and clear the token
	updateData := map[string]any{
		"verified":     true,
		"verify_token": nil,
	}
	if err := db.DB.Model(&user).Updates(updateData).Error; err != nil {
		http.Error(w, "Failed to update The verification Status", http.StatusInternalServerError)
		return
	}

	//parsing the userID stored in string in jwt
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		http.Error(w, "In Valid User details ", http.StatusInternalServerError)
	}

	jwtToken, err := utils.GenerateToken(userID, *user.Name)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return

	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		// Secure:   isProduction,
		// SameSite: http.SameSiteLaxMode,
	})

}
