package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Create the Firebase user with the raw password
	params := (&auth.UserToCreate{}).
		Email(user.Email).
		Password(user.Password). // Use raw password here
		DisplayName(user.Role)   // Assuming Role is used as DisplayName for simplicity

	newUser, err := utils.FirebaseAuth.CreateUser(context.Background(), params)
	if err != nil {
		http.Error(w, "Email already exists", http.StatusInternalServerError)
		log.Printf("Failed to create user: %v\n", err)
		return
	}

	// Hash the password for storage purposes
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		log.Printf("Failed to hash password: %v\n", err)
		return
	}

	// Assign role to the user in Firebase Database
	err = utils.FirebaseDB.NewRef("users/"+newUser.UID+"/role").Set(context.Background(), user.Role)
	if err != nil {
		http.Error(w, "Failed to assign role to user", http.StatusInternalServerError)
		log.Printf("Failed to assign role to user: %v\n", err)
		return
	}

	// Save other user details conditionally
	if user.Gender != "" {
		err = utils.FirebaseDB.NewRef("users/"+newUser.UID+"/gender").Set(context.Background(), user.Gender)
		if err != nil {
			http.Error(w, "Failed to save gender", http.StatusInternalServerError)
			log.Printf("Failed to save gender: %v\n", err)
			return
		}
	}

	if user.PhoneNumber != "" {
		err = utils.FirebaseDB.NewRef("users/"+newUser.UID+"/phone_number").Set(context.Background(), user.PhoneNumber)
		if err != nil {
			http.Error(w, "Failed to save phone number", http.StatusInternalServerError)
			log.Printf("Failed to save phone number: %v\n", err)
			return
		}
	}

	// Save hashed password in Firebase Database
	err = utils.FirebaseDB.NewRef("users/"+newUser.UID+"/hashed_password").Set(context.Background(), string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to save user password", http.StatusInternalServerError)
		log.Printf("Failed to save user password: %v\n", err)
		return
	}

	// Generate a random number between 1 and 10 for profile image assignment
	rand.Seed(time.Now().UnixNano())
	profileImage := rand.Intn(10) + 1 // random number in [1,10]

	// Save the profile image number in Firebase Database
	err = utils.FirebaseDB.NewRef("users/"+newUser.UID+"/profile_image").Set(context.Background(), profileImage)
	if err != nil {
		http.Error(w, "Failed to save profile image", http.StatusInternalServerError)
		log.Printf("Failed to save profile image: %v\n", err)
		return
	}

	// Successfully created user
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully."))
}
