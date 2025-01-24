package controller

import (
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// DeleteAccountRequest defines the request payload structure.
type DeleteAccountRequest struct {
	UID string `json:"uid"` // The UID of the user to be deleted
}

// DeleteAccountHandler handles user account deletion.
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteAccountRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Bad Request: %v\n", err)
		return
	}

	// Ensure UID is provided
	if req.UID == "" {
		http.Error(w, "UID is required", http.StatusBadRequest)
		log.Println("UID is required for account deletion")
		return
	}

	// Delete user from Firebase Authentication
	err := utils.FirebaseAuth.DeleteUser(context.Background(), req.UID)
	if err != nil {
		http.Error(w, "Failed to delete user from authentication", http.StatusInternalServerError)
		log.Printf("Error deleting user from Firebase Auth: %v\n", err)
		return
	}

	// Remove user data from Firebase Realtime Database
	userRef := utils.FirebaseDB.NewRef("users/" + req.UID)
	err = userRef.Delete(context.Background())
	if err != nil {
		http.Error(w, "Failed to delete user data", http.StatusInternalServerError)
		log.Printf("Error deleting user data: %v\n", err)
		return
	}

	// Successfully deleted user
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Account deleted successfully"})
}
