package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	//internals
	"chat-server/internals/db"
	"chat-server/internals/db/models"
	"chat-server/internals/repository"
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

func RegisterHandler(userRepo repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//1.Decoding the Request Body
		var req RegisterRequest
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

		//4.Check the User already exist
		var existingUser models.User
		result := db.DB.Where("name = ? OR email = ?", req.Username, req.Email).First(&existingUser)

		if result.Error == nil {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}
		if result.Error != gorm.ErrRecordNotFound {
			http.Error(w, "Database Error", http.StatusConflict)
			return

		}

		//5.Hash Password
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {

			http.Error(w, "Failed to Hash the Password", http.StatusInternalServerError)
			return
		}

		//6.Verify The Token
		verificationOTP, err := utils.GenerateSecureToken(32)
		if err != nil {
			http.Error(w, "Failed to generateVerification Token", http.StatusInternalServerError)
		}

		hashedPasswordStr := string(passwordHash)
		//7.Create a struct for the new user what to store in the database
		newUser := models.User{
			Name:         &req.Username,
			Email:        &req.Email,
			PasswordHash: &hashedPasswordStr,
			Verified:     false,
			VerifyOTP:    &verificationOTP,
		}

		//8.For the next step using the transcation if we fails nothing will be created.
		err = db.DB.Transaction(func(tx *gorm.DB) error {
			//send verification email to send the otp
			otp, err := SendOTP(*newUser.Email)
			if err != nil {
				return fmt.Errorf("failed to send verification email: %w", err)
			}

			// 2. Save the generated OTP to the user's record.
			newUser.VerifyOTP = &otp

			// 3. Create the user within the transaction.
			if err := tx.Create(&newUser).Error; err != nil {
				return err // GORM errors are returned directly
			}
			return nil
		})

		if err != nil {
			log.Printf("Registration transaction failed: %v", err)
			http.Error(w, "Failed to complete registration", http.StatusInternalServerError)
			return
		}

		//success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		utils.PrettyJSON(w, map[string]any{
			"message": "Registration successful. Please check your email to verify your account.",
			"email":   *newUser.Email,
		})
	}

}
