package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

	// Ensure tags are provided and valid
	if len(post.Tags) == 0 {
		http.Error(w, "At least one tag is required", http.StatusBadRequest)
		return
	}

	for i, tag := range post.Tags {
		post.Tags[i] = strings.TrimSpace(tag) // Normalize tags
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

// GetPostsByTagsHandler fetches posts that match specific tags
func GetPostsByTagsHandler(w http.ResponseWriter, r *http.Request) {
	// Get tags from query parameters
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam == "" {
		http.Error(w, "Tags are required", http.StatusBadRequest)
		return
	}

	// Split the tags by comma and normalize
	tags := strings.Split(tagsParam, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Fetch all posts
	var posts map[string]model.Post
	if err := utils.FirebaseDB.NewRef("posts").Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts by tags
	matchingPosts := make(map[string]model.Post)
	for postID, post := range posts {
		for _, tag := range post.Tags {
			for _, queryTag := range tags {
				if tag == queryTag {
					matchingPosts[postID] = post
					break
				}
			}
		}
	}

	json.NewEncoder(w).Encode(matchingPosts)
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

func FlagPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Retrieve the post
	var post model.Post
	postRef := utils.FirebaseDB.NewRef("posts/" + postID)
	if err := postRef.Get(context.Background(), &post); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Initialize flags map if nil
	if post.Flags == nil {
		post.Flags = make(map[string]bool)
	}

	// If user hasn't already flagged, increase count
	if !post.Flags[request.Username] {
		post.Flags[request.Username] = true
		post.FlagCount++
	}

	// Save back to Firebase
	if err := postRef.Set(context.Background(), post); err != nil {
		http.Error(w, "Failed to flag post", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Post flagged successfully",
		"flag_count": post.FlagCount,
	})
}

// FlagCommentHandler flags a comment by increasing its flag count and storing the username of the flagger
func FlagCommentHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	commentID := r.URL.Query().Get("comment_id")

	if postID == "" || commentID == "" {
		http.Error(w, "Post ID and Comment ID are required", http.StatusBadRequest)
		return
	}

	var request struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Retrieve the post
	var post model.Post
	postRef := utils.FirebaseDB.NewRef("posts/" + postID)
	if err := postRef.Get(context.Background(), &post); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Retrieve the comment
	comment, exists := post.Comments[commentID]
	if !exists {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Initialize flags map if nil
	if comment.Flags == nil {
		comment.Flags = make(map[string]bool)
	}

	// If user hasn't already flagged, increase count
	if !comment.Flags[request.Username] {
		comment.Flags[request.Username] = true
		comment.FlagCount++
	}

	// Save updated comment back into post
	post.Comments[commentID] = comment

	// Save back to Firebase
	if err := postRef.Set(context.Background(), post); err != nil {
		http.Error(w, "Failed to flag comment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Comment flagged successfully",
		"flag_count": comment.FlagCount,
	})
}

// GetFlaggedPostsHandler fetches all posts that have been flagged
func GetFlaggedPostsHandler(w http.ResponseWriter, r *http.Request) {
	var posts map[string]model.Post
	if err := utils.FirebaseDB.NewRef("posts").Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts that have been flagged
	flaggedPosts := make(map[string]model.Post)
	for postID, post := range posts {
		if post.FlagCount > 0 { // Only include flagged posts
			flaggedPosts[postID] = post
		}
	}

	json.NewEncoder(w).Encode(flaggedPosts)
}

// GetFlaggedCommentsHandler fetches all comments that have been flagged
func GetFlaggedCommentsHandler(w http.ResponseWriter, r *http.Request) {
	var posts map[string]model.Post
	if err := utils.FirebaseDB.NewRef("posts").Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	flaggedComments := make(map[string]map[string]model.Comment)

	// Iterate over posts and collect flagged comments
	for postID, post := range posts {
		for commentID, comment := range post.Comments {
			if comment.FlagCount > 0 { // Only include flagged comments
				if _, exists := flaggedComments[postID]; !exists {
					flaggedComments[postID] = make(map[string]model.Comment)
				}
				flaggedComments[postID][commentID] = comment
			}
		}
	}

	json.NewEncoder(w).Encode(flaggedComments)
}

// GetPostsByUsernameHandler fetches all posts created by a specific user
func GetPostsByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Fetch all posts
	var posts map[string]model.Post
	if err := utils.FirebaseDB.NewRef("posts").Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts by username
	userPosts := make(map[string]model.Post)
	for postID, post := range posts {
		if post.Username == username {
			userPosts[postID] = post
		}
	}

	json.NewEncoder(w).Encode(userPosts)
}
