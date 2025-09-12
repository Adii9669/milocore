package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	//internals
	"chat-server/internals/db"
	"chat-server/internals/utils"

	//libraries
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required"`
}

var validate = validator.New()

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	//Decoding the Request Body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body ", http.StatusBadRequest)
		return
	}

	//TrimSpace of the info (username , email ) recieved
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	//validation
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

	//Check the User already exist
	var existingUser db.User
	result := db.DB.Where("name = ? OR email = ?", req.Username, req.Email).First(&existingUser)
	// This condition correctly triggers if a user IS found (result.Error is nil)
	// OR if a real database error occurred.
	if result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	//Hash Password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {

		http.Error(w, "Failed to Hash the Password", http.StatusInternalServerError)
		return
	}

	verificationToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		http.Error(w, "Failed to generateVerification Token", http.StatusInternalServerError)
	}

	//use the transcation if any thing fails the data will not be created in the database

	//Create the User
	newUser := db.User{
		Name:         &req.Username,
		Email:        &req.Email,
		PasswordHash: &[]string{string(passwordHash)}[0],
		Verified:     false,
		VerifyToken:  &verificationToken,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed To Create User", http.StatusInternalServerError)
		return
	}

	//add the user creation
	// creationResult := db.DB.Create(&newUser)
	// if creationResult.Error != nil {
	// 	http.Error(w, "Failed to create the user", http.StatusInternalServerError)
	// 	return
	// }

	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	verificationURL := fmt.Sprintf("%s/api/auth/verify?token=%s", baseURL, verificationToken)
	// verificationURL := "http:/localhost:8000/api/auth/verify?token=" + verificationToken
	err = utils.SendVerificationEmail(*newUser.Email, verificationURL)
	if err != nil {
		log.Printf("CRITICAL: User created (ID: %s) but failed to send verification email: %v", newUser.ID, err)
		//will add the messsaging queue in the production
		http.Error(w, "Failed To Send the Verification Token", http.StatusInternalServerError)
		return
	}

	//success response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful. Please check your email to verify your account."})

}
