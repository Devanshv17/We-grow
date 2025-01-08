package model

// Post represents a user's post
type Post struct {
	ID         string             `json:"id"`
	Username   string             `json:"username"` // Replaced userID with username
	Content    string             `json:"content"`
	ImageURL   string             `json:"image_url"`
	CreatedAt  int64              `json:"created_at"`
	IsResolved bool               `json:"is_resolved"`
	Tags       []string           `json:"tags"`     // New field for tags
	Comments   map[string]Comment `json:"comments"` // Comments stored as a map of Comment structs
}

type Comment struct {
	ID        string `json:"id"`
	Username  string `json:"username"` // Replaced userID with username
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	IsAdmin   bool   `json:"is_admin"`
	Role      string `json:"role"`
}
