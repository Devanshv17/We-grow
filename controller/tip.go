package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"

	"log"
)

// SaveTipHandler allows only admins to post a tip
func SaveTipHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var tip model.Tip
	if err := json.NewDecoder(r.Body).Decode(&tip); err != nil {
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

	// Clear all tips before adding the new one
	err = utils.FirebaseDB.NewRef("tips").Delete(context.Background())
	if err != nil {
		http.Error(w, "Failed to delete existing tips", http.StatusInternalServerError)
		return
	}
	log.Println("All existing tips have been deleted")

	// Push new tip to Firebase
	ref, err := utils.FirebaseDB.NewRef("tips").Push(context.Background(), nil) // Push to generate a unique key
	if err != nil {
		http.Error(w, "Failed to create reference", http.StatusInternalServerError)
		return
	}

	// Set tip data at the generated key
	err = ref.Set(context.Background(), tip)
	if err != nil {
		http.Error(w, "Failed to save tip", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Tip added successfully"))
}

func GetTipsHandler(w http.ResponseWriter, r *http.Request) {
	var tips map[string]model.Tip
	err := utils.FirebaseDB.NewRef("tips").Get(context.Background(), &tips)
	if err != nil {
		http.Error(w, "Failed to retrieve tips", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tips); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
