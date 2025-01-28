package controller

import (
	"backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NotificationRequest struct {
	Topic string `json:"topic"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func CustomNotifHandler(w http.ResponseWriter, r *http.Request) {
	var notification struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("Failed to decode notification body: %v", err)
		return
	}

	err := utils.SendNotificationToTopic("new-videos", notification.Title, notification.Body)
	if err != nil {
		log.Printf("Failed to send notification: %v\n", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Notification sent successfully!")
}
