package controller

import (
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// ChangeUsernameRequest defines the request payload
type ChangeUsernameRequest struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
}

// ChangeUsernameHandler allows users to update their username
func ChangeUsernameHandler(w http.ResponseWriter, r *http.Request) {
	var req ChangeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Bad Request: %v\n", err)
		return
	}

	// Validate request fields
	if req.UID == "" || req.Username == "" {
		http.Error(w, "UID and Username are required", http.StatusBadRequest)
		log.Println("Missing UID or Username in request")
		return
	}

	// Check if the username is unique (if present in the database)
	existingUsers := make(map[string]interface{})
	err := utils.FirebaseDB.NewRef("users").OrderByChild("username").EqualTo(req.Username).Get(context.Background(), &existingUsers)
	if err != nil {
		http.Error(w, "Failed to check username uniqueness", http.StatusInternalServerError)
		log.Printf("Error checking username uniqueness: %v\n", err)
		return
	}

	// If the username is already taken, return a conflict
	if len(existingUsers) > 0 {
		http.Error(w, "Username already exists", http.StatusConflict)
		log.Printf("Username conflict: %s", req.Username)
		return
	}

	// Update the user's username in the database
	userRef := utils.FirebaseDB.NewRef("users/" + req.UID)
	if err := userRef.Update(context.Background(), map[string]interface{}{
		"username": req.Username,
	}); err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		log.Printf("Error updating username: %v\n", err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Username updated successfully"}); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
