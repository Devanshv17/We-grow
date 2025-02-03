// models/video.go
package model

type Video struct {
	Link        string   `json:"link"`
	Tags        []string `json:"tags"`
	Creator     string   `json:"creator"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	IsTopVideo  bool     `json:"isTopVideo"`
	Thumbnail   string   `json:"thumbnail"`
	Rank        int      `json:"rank"`
	Citation    string   `json:"Citation`
}
