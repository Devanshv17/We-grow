package controller

import (
	"backend/model" // Import the Video model from models/video.go
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// SaveVideoHandler handles saving videos to the database
func SaveVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video model.Video
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Bad Request: %v\n", err)
		return
	}

	// Generate a random UUID as the video ID
	videoID := uuid.New().String()

	// Reference to save the video by the random UUID
	videoRef := utils.FirebaseDB.NewRef("videos/" + videoID)
	if err := videoRef.Set(context.Background(), video); err != nil {
		http.Error(w, "Failed to save video", http.StatusInternalServerError)
		log.Printf("Failed to save video: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Video saved successfully"))
}

// GetVideosHandler handles fetching videos from the database for a given user or by tags
func GetVideosHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the UID and tag from the query parameters
	creator := r.URL.Query().Get("creator")
	tag := r.URL.Query().Get("tag")

	var videos map[string]model.Video

	// Fetch all videos from the database
	videosRef := utils.FirebaseDB.NewRef("videos")
	if err := videosRef.Get(context.Background(), &videos); err != nil {
		http.Error(w, "Failed to retrieve videos", http.StatusInternalServerError)
		log.Printf("Failed to retrieve videos: %v\n", err)
		return
	}

	// Filter by Creator if provided
	if creator != "" {
		filteredVideos := make(map[string]model.Video)
		for key, video := range videos {
			if video.Creator == creator { // Ensure video matches the creator
				filteredVideos[key] = video
			}
		}
		videos = filteredVideos
	}

	// Filter by tag if provided
	if tag != "" {
		filteredVideos := make(map[string]model.Video)
		for key, video := range videos {
			for _, videoTag := range video.Tags {
				if videoTag == tag {
					filteredVideos[key] = video
					break
				}
			}
		}
		videos = filteredVideos
	}

	// Return the filtered videos in the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videos); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
