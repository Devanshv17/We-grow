package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// GetProfileHandler handles fetching a user's profile by UID
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the UID from query parameters
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "UID is required", http.StatusBadRequest)
		return
	}

	// Fetch the user's profile from Firebase using the UID
	var user model.User
	userRef := utils.FirebaseDB.NewRef("users/" + uid)
	if err := userRef.Get(context.Background(), &user); err != nil {
		http.Error(w, "Failed to retrieve user profile", http.StatusInternalServerError)
		log.Printf("Failed to retrieve user profile for UID %s: %v\n", uid, err)
		return
	}

	// No need to parse ChildDOB if it's already a string in Firebase
	// Return the user's profile as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
