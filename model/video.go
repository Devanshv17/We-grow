package model

type Video struct {
	Link        string   `json:"link"`
	Tags        []string `json:"tags"`
	Creator     string   `json:"creator"`
	Description string   `json:"description"`
}
