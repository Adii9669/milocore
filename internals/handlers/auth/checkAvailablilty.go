package auth

import (
	"chat-server/internals/db"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type CheckAvailablityRequest struct {
	Field string `json:"field" validate:"required"`
	Value string `json:"value" validate:"required, min=3"`
}

func CheckAvailablityHandler(w http.ResponseWriter, r *http.Request) {
	var req CheckAvailablityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request Body", http.StatusBadRequest)
		return
	}

	req.Field = strings.ToLower(req.Field)
	req.Value = strings.TrimSpace(req.Value)

	var dbColumn string
	// This switch statement is the "brain" of our generic handler.
	// It maps the "field" from the request to the actual database column name.
	switch req.Field {
	case "username":
		dbColumn = "name"
	case "email":
		dbColumn = "email"
	default:
		http.Error(w, fmt.Sprintf("Checking availability for field '%s' is not supported.", req.Field), http.StatusBadRequest)
		return
	}

	// Build the dynamic query
	var count int64
	query := fmt.Sprintf("%s = ?", dbColumn)
	db.DB.Model(&db.User{}).Where(query, req.Value).Count(&count)

	w.Header().Set("Content-Type", "application/json")

	// If the count is 0, the value is available.
	isAvailable := count == 0
	json.NewEncoder(w).Encode(map[string]bool{"available": isAvailable})
}
