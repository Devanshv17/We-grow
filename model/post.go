package model

// Post represents a user's post
type Post struct {
	ID           string             `json:"id"`
	Username     string             `json:"username"` // Replaced userID with username
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	ImageURL     string             `json:"image_url"`
	CreatedAt    int64              `json:"created_at"`
	IsResolved   bool               `json:"is_resolved"`
	Tags         []string           `json:"tags"`     // New field for tags
	Comments     map[string]Comment `json:"comments"` // Comments stored as a map of Comment structs
	Flags        map[string]bool    `json:"flags"`    // Stores usernames who flagged the post
	FlagCount    int                `json:"flag_count"`
	Likes        map[string]bool    `json:"likes"` // Stores usernames who liked the post
	LikeCount    int                `json:"like_count"`
	CommentCount int                `json:"comment_count"` // Counter for comments
}

// Comment represents a comment on a post
type Comment struct {
	ID        string          `json:"id"`
	Username  string          `json:"username"` // Replaced userID with username
	Content   string          `json:"content"`
	CreatedAt int64           `json:"created_at"`
	IsAdmin   bool            `json:"is_admin"`
	Role      string          `json:"role"`
	Flags     map[string]bool `json:"flags"` // Stores usernames who flagged the comment
	FlagCount int             `json:"flag_count"`
	Likes     map[string]bool `json:"likes"` // Stores usernames who liked the comment
	LikeCount int             `json:"like_count"`
}
