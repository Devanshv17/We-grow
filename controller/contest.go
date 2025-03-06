package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// SaveContestHandler allows only admins to update the contest
func SaveContestHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var contest model.Contest
	if err := json.NewDecoder(r.Body).Decode(&contest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user_id from request
	userID := r.Header.Get("user_id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Retrieve user role
	var role string
	err := utils.FirebaseDB.NewRef("users/"+userID+"/role").Get(context.Background(), &role)
	if err != nil || role != "admin" {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	log.Printf("Retrieved role: %s", role)

	// Clear existing contest before adding the new one
	err = utils.FirebaseDB.NewRef("contest").Delete(context.Background())
	if err != nil {
		http.Error(w, "Failed to delete existing contest", http.StatusInternalServerError)
		return
	}
	log.Println("Existing contest data has been deleted")

	// Save new contest details
	err = utils.FirebaseDB.NewRef("contest").Set(context.Background(), contest)
	if err != nil {
		http.Error(w, "Failed to save contest details", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Contest updated successfully"))
}

// GetContestHandler retrieves the current contest details
func GetContestHandler(w http.ResponseWriter, r *http.Request) {
	var contest model.Contest
	err := utils.FirebaseDB.NewRef("contest").Get(context.Background(), &contest)
	if err != nil {
		http.Error(w, "Failed to retrieve contest details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(contest); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
