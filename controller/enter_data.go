package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// EnterDataRequest structure for the request body
type EnterDataRequest struct {
	UID          string `json:"uid"`
	PhoneNumber  string `json:"phone_number,omitempty"` // Make phone number optional
	Name         string `json:"name"`
	Gender       string `json:"gender"` // 'male', 'female', 'others'
	City         string `json:"city"`
	ChildDOB     string `json:"child_dob"`
	ProfileImage int    `json:"profile_image"` // Optional field for profile image number (1-10)
}

// EnterDataHandler function to update user data
func EnterDataHandler(w http.ResponseWriter, r *http.Request) {
	var req EnterDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Bad Request: %v\n", err)
		return
	}

	// Validate gender
	if req.Gender != "male" && req.Gender != "female" && req.Gender != "others" {
		http.Error(w, "Invalid gender option", http.StatusBadRequest)
		log.Println("Invalid gender option:", req.Gender)
		return
	}

	// If the phone number is provided and it needs to be updated, check if it's unique
	if req.PhoneNumber != "" {
		existingUsers := make(map[string]model.User) // Use a map to hold existing users
		err := utils.FirebaseDB.NewRef("users").OrderByChild("phone_number").EqualTo(req.PhoneNumber).Get(context.Background(), &existingUsers)
		if err != nil {
			http.Error(w, "Failed to check phone number uniqueness", http.StatusInternalServerError)
			log.Printf("Failed to check phone number uniqueness: %v\n", err)
			return
		}

		// If a user with the specified phone number exists, return a conflict status
		if len(existingUsers) > 0 {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			log.Println("Phone number already exists:", req.PhoneNumber)
			return
		}
	}

	// Retrieve the UID from the request
	uid := req.UID

	// Create a map to update only non-empty fields
	updateData := map[string]interface{}{
		"name":      req.Name,
		"gender":    req.Gender,
		"city":      req.City,
		"child_dob": req.ChildDOB,
	}

	// If a new phone number is provided, include it in the update data
	if req.PhoneNumber != "" {
		updateData["phone_number"] = req.PhoneNumber
	}

	// If a profile image number is provided (non-zero), validate and include it in the update data
	if req.ProfileImage != 0 {
		if req.ProfileImage < 1 || req.ProfileImage > 10 {
			http.Error(w, "Invalid profile image value; must be between 1 and 10", http.StatusBadRequest)
			return
		}
		updateData["profile_image"] = req.ProfileImage
	}

	// Update the user's details in Firebase Database
	userRef := utils.FirebaseDB.NewRef("users/" + uid)
	if err := userRef.Update(context.Background(), updateData); err != nil {
		http.Error(w, "Failed to update user data", http.StatusInternalServerError)
		log.Printf("Error updating user data: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "User data updated successfully"}); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
		return
	}
}
