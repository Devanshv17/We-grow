package controller

import (
	"backend/model" // Import the Video model from models/video.go
	"backend/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"

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

	// Send a customized notification for the video
	title := "New Video Posted: " + video.Title
	body := "Check out " + video.Creator + "'s latest video on " + video.Title + "!"

	err := utils.SendNotificationToTopic("new-videos", title, body)
	if err != nil {
		log.Printf("Failed to send notification: %v\n", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("Video saved successfully and notification sent for video ID: %s", videoID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Video saved and notification sent successfully"))
}

// GetVideosHandler handles fetching videos from the database for a given user or by tags
func GetVideosHandler(w http.ResponseWriter, r *http.Request) {
	creator := r.URL.Query().Get("creator")
	tag := r.URL.Query().Get("tag")

	var videos map[string]model.Video

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
			if video.Creator == creator {
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

	// Convert map to slice for sorting
	videoList := make([]model.Video, 0, len(videos))
	for _, video := range videos {
		videoList = append(videoList, video)
	}

	// Sort videos in descending order of rank (higher rank first)
	sort.Slice(videoList, func(i, j int) bool {
		return videoList[i].Rank > videoList[j].Rank
	})

	// Return the sorted videos
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videoList); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Printf("Failed to encode response: %v\n", err)
	}
}
