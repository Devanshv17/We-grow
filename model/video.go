package model

type Video struct {
	Link        string   `json:"link"`
	Tags        []string `json:"tags"`
	Creator     string   `json:"creator"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
}
