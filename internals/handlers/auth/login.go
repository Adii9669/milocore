package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	// internals
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
	"chat-server/internals/requests"
	"chat-server/internals/utils"

	// libraries
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// helper Function to write the json errors
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// email validation
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%^&*+\\/?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isEmail(e string) bool {
	return emailRegex.MatchString(e)
}

func LoginHandler(userRepo repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req requests.Credentials
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid request Body.", http.StatusBadRequest)
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeJSONError(w, "Check the Empty Fields.", http.StatusBadRequest)
			return
		}

		var user *models.User
		var err error
		if isEmail(req.Username) {
			user, err = userRepo.FindByEmail(ctx, req.Username)
		} else {
			user, err = userRepo.FindBYName(ctx, req.Username)
		}
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				writeJSONError(w, "Invalid Credentials", http.StatusUnauthorized)
			} else {
				writeJSONError(w, "Database Error", http.StatusInternalServerError)
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
			writeJSONError(w, "Invalid Email OR Password", http.StatusUnauthorized)
			return
		}

		if !user.Verified {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":  "Account not verified. Please check your email to complete registration.",
				"status": "unverified",
			})
			return
		}

		userID, err := uuid.Parse(user.ID.String())
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
			return
		}

		tokenString, err := utils.GenerateToken(userID, user.Email)
		if err != nil {
			writeJSONError(w, "Invalid Token", http.StatusInternalServerError)
			return
		}

		isProduction := os.Getenv("APP_ENV") == "production"
		sameSite := http.SameSiteLaxMode
		secure := false
		if isProduction {
			secure = true
			sameSite = http.SameSiteNoneMode
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   int(utils.TokenExpiryDuration.Seconds()),
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		utils.PrettyJSON(w, map[string]any{
			"message": "Logged In Successfully",
			"user": map[string]any{
				"id":     user.ID,
				"name":   user.Name,
				"email":  user.Email,
				"status": "verified",
			},
		})
	}
}
