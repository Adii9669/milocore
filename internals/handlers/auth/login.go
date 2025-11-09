package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	//internals
	"chat-server/internals/db/models"
	"chat-server/internals/handlers"
	"chat-server/internals/repository"
	"chat-server/internals/utils"

	//libraries
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
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%^&*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func isEmail(e string) bool {
	return emailRegex.MatchString(e)
}

func LoginHandler(userRepo repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//1.Decode the incoming request
		var req handlers.Credentials
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid request Body.", http.StatusBadRequest)
			return
		}

		//2.validate the data if empty or not
		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" || req.Password == "" {
			writeJSONError(w, "Check the Empty Feilds.", http.StatusBadRequest)
			return
		}

		//3.find the user or the email present in the database or not
		var user *models.User
		var err error

		if isEmail(req.Username) {
			user, err = userRepo.FindByEmail(req.Username)
		} else {
			user, err = userRepo.FindBYName(req.Username)
		}
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				writeJSONError(w, "Invalid Credentials", http.StatusUnauthorized)
			} else {
				writeJSONError(w, "Database Error", http.StatusInternalServerError)
			}
			return
		}

		//4. Compare Hashed Password
		if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
			writeJSONError(w, "Invalid Email OR Passoword", http.StatusUnauthorized)
			return
		}

		//5.check if the user is verified or not
		if !user.Verified {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error":  "Account not verified. Please check your email to complete registration.",
				"status": "unverified",
			})
			return
		}

		//helper function for the parsing the uuid which is stored in string type in jwt
		userID, err := uuid.Parse(user.ID)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusInternalServerError)
			return
		}

		//6.generate the token for that user
		tokenString, err := utils.GenerateToken(userID, *user.Email)
		if err != nil {
			writeJSONError(w, "Invalid Token", http.StatusInternalServerError)
		}

		//only for production to secure the connection
		isProduction := os.Getenv("APP_ENV") == "production"
		sameSite := http.SameSiteLaxMode
		secure := false

		if isProduction {
			// local dev, no HTTPS normally
			sameSite = http.SameSiteNoneMode
			secure = true
		}
		//7. Set and secure the token in the cookies
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Path:     "/",
			MaxAge:   3600,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})

		//7.Login Complete send back the response with the details
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
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
