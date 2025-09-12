package auth

import (
	"chat-server/internals/db"
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)
type AvailableRequest struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type AvailabilityResponse struct {
	Available bool `json:"available"`
}

func CheckAvailablityHandler(w http.ResponseWriter, r *http.Request) {
	var req AvailableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	var existingUser db.User

	result := db.DB.Where("name = ?", req.Value).First(&existingUser)

	// If gorm.ErrRecordNotFound, it means the user was NOT found, so the name is available.
	available := result.Error == gorm.ErrRecordNotFound

	// send the simple true/false response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AvailabilityResponse{Available: available})

}
