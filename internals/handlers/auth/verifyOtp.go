package auth

import (
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/handlers"
	"chat-server/internals/repository"
	"chat-server/internals/utils"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"

	"gorm.io/gorm"
)

func VerifyOtpHandler(userRepo repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//take the request and check it
		var req handlers.VerifyOtpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		//request verified now check the details in it
		user, err := userRepo.FindByEmail(req.Email)
		log.Printf("Finding the user %v", user)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		//now next check if he is already verified
		if user.Verified {
			http.Error(w, "Account is already verified", http.StatusBadRequest)
			return
		}

		// 3. Check if the submitted OTP matches the one in the database.
		if user.VerifyOTP == nil || *user.VerifyOTP != req.OTP {
			http.Error(w, "Invalid verification code", http.StatusBadRequest)
			return
		}

		// 4. Update the user: mark as verified and clear the token.
		user.Verified = true
		user.VerifyOTP = nil // Clear the token so it can't be used again

		if err := db.DB.Save(&user).Error; err != nil {
			http.Error(w, "Failed to update user status", http.StatusInternalServerError)
			return
		}

		//parsing the uuid which is stored in string type in jwt
		userID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
			return
		}

		// 5. Generate a JWT using your utility function.
		tokenString, err := utils.GenerateToken(userID, *user.Name) // Use your function
		if err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}

		// 6. Set the secure HttpOnly cookie for the browser.
		isProduction := config.Cfg.CHECK_ENV.ENV == "production"
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   3600 * 24, // 24 hours
			HttpOnly: true,
			Secure:   isProduction,
			SameSite: http.SameSiteLaxMode,
		})

		// 7. Send a success response with the new token.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{
			"message":  "Account verified successfully.",
			"Verified": user.Verified,
		})

	}
}
