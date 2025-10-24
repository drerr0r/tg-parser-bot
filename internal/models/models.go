package models

import "time"

type Post struct {
	ID          string    `json:"id" bson:"_id"`
	Source      string    `json:"source"`
	Content     string    `json:"content"`
	MediaURLs   []string  `json:"media_urls,omitempty"`
	Author      string    `json:"author,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Published   bool      `json:"published"`
	PublishedTo []string  `json:"published_to,omitempty"` // ["tg", "vk"]
}

type Config struct {
	Telegram struct {
		APIID    int    `yaml:"api_id"`
		APIHash  string `yaml:"api_hash"`
		Phone    string `yaml:"phone"`
		Channels []struct {
			Username string `yaml:"username"`
			LastID   int    `yaml:"last_id"`
		} `yaml:"channels"`
		TargetChannel string `yaml:"target_channel"`
	} `yaml:"telegram"`

	VK struct {
		AccessToken string `yaml:"access_token"`
		GroupID     int    `yaml:"group_id"`
		AlbumID     int    `yaml:"album_id,omitempty"`
	} `yaml:"vk"`

	Database struct {
		MongoDBURI string `yaml:"mongo_uri"`
		Database   string `yaml:"database"`
		Collection string `yaml:"collection"`
	} `yaml:"database"`

	App struct {
		CheckInterval  int `yaml:"check_interval"` // в секундах
		MaxPostsPerRun int `yaml:"max_posts_per_run"`
	} `yaml:"app"`
}
