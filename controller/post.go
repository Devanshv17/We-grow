package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// CreatePostHandler creates a new post and stores it in the database
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post model.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Bad Request: Unable to decode JSON", http.StatusBadRequest)
		return
	}

	// Ensure that the username is provided
	if post.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Set post metadata
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now().Unix()
	post.IsResolved = false

	// Save post to Firebase
	postRef := utils.FirebaseDB.NewRef("posts/" + post.ID)
	if err := postRef.Set(context.Background(), post); err != nil {
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(post)
}

// AddCommentHandler adds a comment to a specific post
func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment model.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Bad Request: Unable to decode JSON", http.StatusBadRequest)
		return
	}

	// Check if username is empty or missing
	if comment.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Query to get the user by username
	var users map[string]model.User
	err := utils.FirebaseDB.NewRef("users").
		OrderByChild("username").
		EqualTo(comment.Username).
		LimitToFirst(1).
		Get(context.Background(), &users)

	if err != nil || len(users) == 0 {
		http.Error(w, "Failed to verify user role", http.StatusInternalServerError)
		return
	}

	// Get the first user (should be unique due to username being unique)
	var user model.User
	for _, u := range users {
		user = u
		break
	}

	// Set comment metadata
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now().Unix()
	comment.IsAdmin = user.Role == "admin"
	comment.Role = user.Role

	// Get post ID from query parameters
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Save the comment directly inside the post under the "comments" field
	commentRef := utils.FirebaseDB.NewRef("posts/" + postID + "/comments/" + comment.ID)
	if err := commentRef.Set(context.Background(), comment); err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comment)
}

// GetPostsHandler fetches all posts and optionally includes comments
func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	var posts map[string]model.Post
	if err := utils.FirebaseDB.NewRef("posts").Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Check if comments should be included (optional query param)
	includeComments := r.URL.Query().Get("includeComments") == "true"
	if includeComments {
		for postID, post := range posts {
			// Fetch the comments for each post directly from Firebase
			var comments map[string]model.Comment
			err := utils.FirebaseDB.NewRef("posts/"+postID+"/comments").Get(context.Background(), &comments)
			if err != nil {
				http.Error(w, "Failed to fetch comments for post "+postID, http.StatusInternalServerError)
				return
			}

			// Directly assign the fetched comments (no need to convert to a slice)
			post.Comments = comments
			posts[postID] = post
		}
	}

	// Return the posts, now including comments if requested
	json.NewEncoder(w).Encode(posts)
}
