package utils

import (
	"context"
	"errors"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"firebase.google.com/go/db"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var (
	FirebaseAuth *auth.Client
	FirebaseDB   *db.Client
)

func InitFirebase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}

	opt := option.WithCredentialsFile("firebase.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v\n", err)
	}

	FirebaseAuth, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth client: %v\n", err)
	}

	// Get the database URL from the environment variables
	databaseURL := os.Getenv("FIREBASE_DATABASE_URL")
	if databaseURL == "" {
		log.Fatalf("FIREBASE_DATABASE_URL not set in .env file")
	}

	// Initialize the Firebase Database client with the database URL
	FirebaseDB, err = app.DatabaseWithURL(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Error initializing Firebase Database client: %v\n", err)
	}
}

func VerifyIDToken(idToken string) (*auth.Token, error) {
	// Verify the token using the Firebase Auth client
	token, err := FirebaseAuth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		log.Printf("Error verifying ID token: %v", err)
		return nil, errors.New("Invalid or expired token")
	}

	// Return the decoded token (which contains the user's UID and other claims)
	return token, nil
}
