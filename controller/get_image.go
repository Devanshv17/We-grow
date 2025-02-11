package controller

import (
	"backend/utils" // Make sure this package exposes your initialized Firebase Database (FirebaseDB)
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// GetProfileImageHandler retrieves the profile image number for a given username.
func GetProfileImageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the "username" parameter from the query string.
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username parameter", http.StatusBadRequest)
		return
	}

	// Reference the "users" node.
	ref := utils.FirebaseDB.NewRef("users")

	// Query for the user with the specified username.
	var users map[string]map[string]interface{}
	err := ref.OrderByChild("username").EqualTo(username).Get(context.Background(), &users)
	if err != nil {
		log.Printf("Error querying user by username %s: %v", username, err)
		http.Error(w, "Error querying user", http.StatusInternalServerError)
		return
	}

	// Check if any user was found.
	if users == nil || len(users) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Retrieve the profile_image field from the first matching user.
	var profileImage interface{}
	for _, user := range users {
		if v, ok := user["profile_image"]; ok {
			profileImage = v
			break
		}
	}

	if profileImage == nil {
		http.Error(w, "Profile image not set for user", http.StatusNotFound)
		return
	}

	// Prepare the JSON response.
	response := map[string]interface{}{
		"username":      username,
		"profile_image": profileImage,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
