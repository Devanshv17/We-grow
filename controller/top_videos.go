// Add the new handler in the controller package
package controller

import (
	"backend/model"
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"
)

// SaveTopVideoHandler handles saving top videos to the database
func SaveTopVideoHandler(w http.ResponseWriter, r *http.Request) {
	var video model.Video
	if err := json.NewDecoder(r.Body).Decode(&video); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Bad Request: %v\n", err)
		return
	}

	// Generate a random UUID as the video ID
	videoID := uuid.New().String()

	// Set the IsTopVideo attribute to true
	video.IsTopVideo = true

	// Reference to save the video as a top video by the random UUID
	topVideoRef := utils.FirebaseDB.NewRef("top_videos/" + videoID)
	if err := topVideoRef.Set(context.Background(), video); err != nil {
		http.Error(w, "Failed to save top video", http.StatusInternalServerError)
		log.Printf("Failed to save top video: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Top video saved successfully"))
}

// GetTopVideosHandler handles fetching top videos from the database
func GetTopVideosHandler(w http.ResponseWriter, r *http.Request) {
	var topVideos map[string]model.Video

	// Fetch all top videos from the database
	topVideosRef := utils.FirebaseDB.NewRef("top_videos")
	if err := topVideosRef.Get(context.Background(), &topVideos); err != nil {
		http.Error(w, "Failed to retrieve top videos", http.StatusInternalServerError)
		log.Printf("Failed to retrieve top videos: %v\n", err)
		return
	}

	// Convert map to slice for sorting
	videoList := make([]model.Video, 0, len(topVideos))
	for _, video := range topVideos {
		videoList = append(videoList, video)
	}

	// Sort videos in descending order of rank (higher rank first)
	sort.Slice(videoList, func(i, j int) bool {
		return videoList[i].Rank > videoList[j].Rank
	})

	// Return the sorted top videos
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videoList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
