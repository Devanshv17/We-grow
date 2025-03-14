package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	// Use UnixNano (and negate it) to get a high-precision timestamp for sorting from new to older.
	post.CreatedAt = -time.Now().UnixNano()
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

	// Split and normalize tags
	tags := strings.Split(tagsParam, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	// Set default limit to 4; override if provided in query params.
	limit := 4
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil {
			limit = l
		} else {
			log.Println("Invalid limit parameter:", limitParam)
		}
	}

	// Retrieve starting timestamp for pagination if provided.
	startAfterParam := r.URL.Query().Get("startAfter")

	// Begin Firebase query ordered by "created_at".
	ref := utils.FirebaseDB.NewRef("posts")
	query := ref.OrderByChild("created_at")

	// If a valid startAfter parameter is provided, update the query.
	if startAfterParam != "" {
		if startAfter, err := strconv.ParseInt(startAfterParam, 10, 64); err == nil {
			// Add 1 to the timestamp to avoid including the last fetched post again.
			query = query.StartAt(startAfter + 1)
		} else {
			log.Println("Invalid startAfter parameter:", startAfterParam)
		}
	}

	// Limit the query to the desired number of posts.
	query = query.LimitToFirst(limit)

	// Execute the query.
	var posts map[string]model.Post
	if err := query.Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts by matching tags.
	matchingPosts := make(map[string]model.Post)
	for postID, post := range posts {
		for _, postTag := range post.Tags {
			for _, queryTag := range tags {
				if postTag == queryTag {
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

	// Retrieve post to update CommentCount
	var post model.Post
	postRef := utils.FirebaseDB.NewRef("posts/" + postID)
	if err := postRef.Get(context.Background(), &post); err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Save the comment directly inside the post under the "comments" field
	commentRef := utils.FirebaseDB.NewRef("posts/" + postID + "/comments/" + comment.ID)
	if err := commentRef.Set(context.Background(), comment); err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Increment the CommentCount
	post.CommentCount++

	if err := postRef.Update(context.Background(), map[string]interface{}{
		"comment_count": post.CommentCount,
	}); err != nil {
		log.Println("Failed to update comment count:", err)
		http.Error(w, "Failed to update comment count", http.StatusInternalServerError)
		return
	}

	// Save updated post back to Firebase
	// if err := postRef.Set(context.Background(), post); err != nil {
	// 	http.Error(w, "Failed to update comment count", http.StatusInternalServerError)
	// 	return
	// }

	json.NewEncoder(w).Encode(comment)
}

func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Set a default limit of 5 posts.
	limit := 4
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil {
			limit = l
		} else {
			log.Println("Invalid limit parameter:", limitParam)
		}
	}

	// Retrieve the starting timestamp for pagination (if provided).
	startAfterParam := r.URL.Query().Get("startAfter")

	// Begin the Firebase query ordering by the "created_at" field.
	ref := utils.FirebaseDB.NewRef("posts")
	query := ref.OrderByChild("created_at")

	// If a valid startAfter parameter is provided, update the query.
	if startAfterParam != "" {
		if startAfter, err := strconv.ParseInt(startAfterParam, 10, 64); err == nil {
			// Add 1 to the timestamp to avoid including the last fetched post again.
			query = query.StartAt(startAfter + 1)
		} else {
			log.Println("Invalid startAfter parameter:", startAfterParam)
		}
	}

	// Limit the query to the desired number of posts.
	query = query.LimitToFirst(limit)

	// Execute the query.
	var posts map[string]model.Post
	if err := query.Get(context.Background(), &posts); err != nil {
		log.Println("Error fetching posts:", err)
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Optionally include comments if requested.
	includeComments := r.URL.Query().Get("includeComments") == "true"
	if includeComments && posts != nil {
		for postID, post := range posts {
			var comments map[string]model.Comment
			if err := utils.FirebaseDB.NewRef("posts/"+postID+"/comments").Get(context.Background(), &comments); err != nil {
				log.Println("Error fetching comments for post", postID, ":", err)
				http.Error(w, "Failed to fetch comments for post "+postID, http.StatusInternalServerError)
				return
			}
			post.Comments = comments
			posts[postID] = post
		}
	}

	// Return the fetched posts as JSON.
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

// GetPostsByUsernameHandler fetches all posts created by a specific user,
// with pagination and ordering by created_at (default limit is 4).
func GetPostsByUsernameHandler(w http.ResponseWriter, r *http.Request) {
	// Get username from query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Set default limit to 4; override if provided in query params.
	limit := 4
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil {
			limit = l
		} else {
			log.Println("Invalid limit parameter:", limitParam)
		}
	}

	// Retrieve starting timestamp for pagination if provided.
	startAfterParam := r.URL.Query().Get("startAfter")

	// Begin Firebase query ordered by "created_at".
	ref := utils.FirebaseDB.NewRef("posts")
	query := ref.OrderByChild("created_at")

	// If a valid startAfter parameter is provided, update the query.
	if startAfterParam != "" {
		if startAfter, err := strconv.ParseInt(startAfterParam, 10, 64); err == nil {
			// Add 1 to the timestamp to avoid including the last fetched post again.
			query = query.StartAt(startAfter + 1)
		} else {
			log.Println("Invalid startAfter parameter:", startAfterParam)
		}
	}

	// Limit the query to the desired number of posts.
	query = query.LimitToFirst(limit)

	// Execute the query.
	var posts map[string]model.Post
	if err := query.Get(context.Background(), &posts); err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	// Filter posts by username.
	userPosts := make(map[string]model.Post)
	for postID, post := range posts {
		if post.Username == username {
			userPosts[postID] = post
		}
	}

	json.NewEncoder(w).Encode(userPosts)
}

// LikeCommentHandler handles liking and unliking a comment
func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Retrieve the comment
	var comment model.Comment
	commentRef := utils.FirebaseDB.NewRef("posts/" + postID + "/comments/" + commentID)
	if err := commentRef.Get(context.Background(), &comment); err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Initialize likes map if nil
	if comment.Likes == nil {
		comment.Likes = make(map[string]bool)
	}

	// Toggle like status
	if comment.Likes[request.Username] {
		delete(comment.Likes, request.Username)
		comment.LikeCount--
	} else {
		comment.Likes[request.Username] = true
		comment.LikeCount++
	}

	// Save back to Firebase
	if err := commentRef.Set(context.Background(), comment); err != nil {
		http.Error(w, "Failed to update like status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Like status updated",
		"like_count": comment.LikeCount,
	})
}

// LikePostHandler handles liking and unliking a post
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
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

	// Initialize likes map if nil
	if post.Likes == nil {
		post.Likes = make(map[string]bool)
	}

	// Toggle like status
	if post.Likes[request.Username] {
		delete(post.Likes, request.Username)
		post.LikeCount--
	} else {
		post.Likes[request.Username] = true
		post.LikeCount++
	}

	// Save back to Firebase
	if err := postRef.Set(context.Background(), post); err != nil {
		http.Error(w, "Failed to update like status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Like status updated",
		"like_count": post.LikeCount,
	})
}
