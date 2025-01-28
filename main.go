package main

import (
	"backend/controller"
	"backend/middleware"
	"backend/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize Firebase Auth and Database clients
	utils.InitFirebase()

	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(middleware.CORS)
	// Register routes
	r.HandleFunc("/register", controller.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", controller.LoginHandler).Methods("POST")
	r.HandleFunc("/forget-password", controller.ForgotPasswordHandler).Methods("POST")
	r.HandleFunc("/delete_account", controller.DeleteAccountHandler).Methods("POST")
	r.HandleFunc("/resend-verification", controller.ResendVerificationHandler).Methods("POST")
	r.HandleFunc("/enter_data", controller.EnterDataHandler).Methods("POST")
	r.HandleFunc("/username", controller.ChangeUsernameHandler).Methods("POST")
	r.HandleFunc("/videos", controller.SaveVideoHandler).Methods("POST")
	r.HandleFunc("/videos", controller.GetVideosHandler).Methods("GET")
	r.HandleFunc("/videos/top", controller.SaveTopVideoHandler).Methods("POST")
	r.HandleFunc("/videos/top", controller.GetTopVideosHandler).Methods("GET")
	r.HandleFunc("/profile", controller.GetProfileHandler).Methods("GET")
	r.HandleFunc("/posts", controller.CreatePostHandler).Methods("POST")
	r.HandleFunc("/posts", controller.GetPostsHandler).Methods("GET")
	r.HandleFunc("/posts/flag", controller.FlagPostHandler).Methods("POST")
	r.HandleFunc("/comments/flag", controller.FlagCommentHandler).Methods("POST")
	r.HandleFunc("/posts/comment", controller.AddCommentHandler).Methods("POST")
	r.HandleFunc("/posts/tags", controller.GetPostsByTagsHandler).Methods("GET")
	r.HandleFunc("/custom-notif", controller.CustomNotifHandler).Methods("POST")

	// Start server

	fmt.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
