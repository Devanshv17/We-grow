package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Authenticate user with Firebase Authentication
	u, err := utils.FirebaseAuth.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		log.Printf("Failed to get user: %v\n", err)
		return
	}

	if !u.EmailVerified {
		http.Error(w, "Email not verified", http.StatusUnauthorized)
		log.Printf("Email not verified for user: %v\n", user.Email)
		return
	}

	// Retrieve user's role from Firebase Database
	var role string
	err = utils.FirebaseDB.NewRef("users/"+u.UID+"/role").Get(context.Background(), &role)
	if err != nil || role == "" {
		http.Error(w, "Failed to retrieve user role", http.StatusInternalServerError)
		log.Printf("Failed to retrieve user role: %v\n", err)
		return
	}

	// Create the response payload including UID and role
	response := map[string]interface{}{
		"user_id": u.UID, // Use the UID directly instead of email
		"role":    role,
	}

	// Return the user ID and role in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
