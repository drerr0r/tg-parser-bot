package models

import (
	"time"
)

// PlatformType - тип платформы
type PlatformType string

const (
	PlatformTelegram PlatformType = "telegram"
	PlatformVK       PlatformType = "vk"
)

// MediaType - тип медиа контента
type MediaType string

const (
	MediaText     MediaType = "text"
	MediaPhoto    MediaType = "photo"
	MediaVideo    MediaType = "video"
	MediaDocument MediaType = "document"
	MediaVoice    MediaType = "voice"
	MediaSticker  MediaType = "sticker"
)

// ParsingRule - правило парсинга
type ParsingRule struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	SourceChannel    string            `json:"source_channel"`
	Keywords         []string          `json:"keywords"`
	ExcludeWords     []string          `json:"exclude_words"`
	MediaTypes       []MediaType       `json:"media_types"`
	MinTextLength    int               `json:"min_text_length"`
	MaxTextLength    int               `json:"max_text_length"`
	TextReplacements map[string]string `json:"text_replacements"`
	AddPrefix        string            `json:"add_prefix"`
	AddSuffix        string            `json:"add_suffix"`
	TargetPlatforms  []PlatformType    `json:"target_platforms"`
	CheckInterval    int               `json:"check_interval"`
	IsActive         bool              `json:"is_active"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// Post - модель поста
type Post struct {
	ID            int64     `json:"id"`
	RuleID        int64     `json:"rule_id"`
	MessageID     int64     `json:"message_id"`
	SourceChannel string    `json:"source_channel"`
	Content       string    `json:"content"`
	MediaType     MediaType `json:"media_type"`
	MediaURL      string    `json:"media_url"`
	PostedAt      time.Time `json:"posted_at"`
	ParsedAt      time.Time `json:"parsed_at"`

	PublishError      string `json:"publish_error"`
	PublishedTelegram bool   `json:"published_telegram" db:"published_telegram"`
	PublishedVK       bool   `json:"published_vk" db:"published_vk"`
}

// Config - основная структура конфигурации
type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Telegram TelegramConfig `yaml:"telegram"`
	VK       VKConfig       `yaml:"vk"`
	Server   ServerConfig   `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	Auth     AuthConfig     `yaml:"auth"`
}

// AuthConfig конфигурация аутентификации
type AuthConfig struct {
	JWTSecret   string `yaml:"jwt_secret"`
	JWTDuration int    `yaml:"jwt_duration"` // в часах
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type TelegramConfig struct {
	APIID         int    `yaml:"api_id"`
	APIHash       string `yaml:"api_hash"`
	Phone         string `yaml:"phone"`
	SessionFile   string `yaml:"session_file"`
	BotToken      string `yaml:"bot_token"`
	TargetChannel string `yaml:"target_channel"`
}

type VKConfig struct {
	AccessToken string `yaml:"access_token"`
	GroupID     int    `yaml:"group_id"`
	Version     string `yaml:"version"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LoggerConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	FilePath string `yaml:"file_path"`
}
