package model

type Video struct {
	Link string   `json:"link"`
	UID  string   `json:"uid"`
	Tags []string `json:"tags"`
}
