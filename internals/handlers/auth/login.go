package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	//internals
	"chat-server/internals/db"
	"chat-server/internals/utils"

	//libraries
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

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	//Decode the incoming request
	var req Credentials
	//take json and decode it into the Credentials
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request Body.", http.StatusBadRequest)
		return
	}

	//validate the data
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeJSONError(w, "Check the Empty Feilds.", http.StatusBadRequest)
		return
	}

	//find the user or the email
	var user db.User
	var err error

	if isEmail(req.Username) {
		err = db.DB.Where("email = ?", req.Username).First(&user).Error
	} else {
		err = db.DB.Where("name = ?", req.Username).First(&user).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			writeJSONError(w, "Invalid Credentials", http.StatusUnauthorized)
		} else {
			writeJSONError(w, "Database Error", http.StatusInternalServerError)
		}
		return
	}

	// CompareHashAndPassword
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSONError(w, "Invalid Email OR Passoword", http.StatusUnauthorized)
		return
	}

	//check if the user is verified or not
	if !user.Verified {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  "Account not verified. Please check your email to complete registration.",
			"status": "unverified",
		})
		return
	}

	//generate the token for that user
	tokenString, err := utils.GenerateToken(user.ID, *user.Email)
	if err != nil {
		writeJSONError(w, "Invalid Token", http.StatusInternalServerError)
	}

	//only for production to secure the connection
	isProduction := os.Getenv("APP_ENV") == "production"
	//secure the token in the session
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	//Login Complete
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
