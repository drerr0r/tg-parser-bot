package models

import (
	"errors"
	"time"
)

// NewPost создает новый пост
func NewPost(ruleID, messageID int64, sourceChannel, content string, mediaType MediaType) *Post {
	now := time.Now()
	return &Post{
		RuleID:            ruleID,
		MessageID:         messageID,
		SourceChannel:     sourceChannel,
		Content:           content,
		MediaType:         mediaType,
		PostedAt:          now,
		ParsedAt:          now,
		PublishedTelegram: false,
		PublishedVK:       false,
		PublishError:      "",
	}
}

// MarkAsPublished помечает пост как опубликованный
func (p *Post) MarkAsPublished() {
	p.PublishedTelegram = true
	p.PublishedVK = true
	p.PublishError = ""
}

// MarkAsFailed помечает пост как неопубликованный с ошибкой
func (p *Post) MarkAsFailed(errorMsg string) {
	p.PublishedTelegram = false
	p.PublishedVK = false
	p.PublishError = errorMsg
}

// SetMediaURL устанавливает URL медиа
func (p *Post) SetMediaURL(url string) {
	p.MediaURL = url
}

// IsProcessed проверяет, был ли пост обработан
func (p *Post) IsProcessed() bool {
	return p.PublishedTelegram || p.PublishedVK || p.PublishError != ""
}

// CanRetry проверяет, можно ли повторно отправить пост
func (p *Post) CanRetry() bool {
	return (!p.PublishedTelegram || !p.PublishedVK) && p.PublishError != ""
}

// Validate проверяет валидность поста
func (p *Post) Validate() error {
	if p.RuleID == 0 {
		return errors.New("rule ID is required")
	}
	if p.MessageID == 0 {
		return errors.New("message ID is required")
	}
	if p.SourceChannel == "" {
		return errors.New("source channel is required")
	}
	if p.Content == "" && p.MediaType == MediaText {
		return errors.New("content is required for text posts")
	}
	return nil
}

// GetSummary возвращает краткое описание поста
func (p *Post) GetSummary() string {
	contentPreview := p.Content
	if len(contentPreview) > 100 {
		contentPreview = contentPreview[:100] + "..."
	}

	status := "Ожидает"
	if p.PublishedTelegram && p.PublishedVK {
		status = "Опубликован везде"
	} else if p.PublishedTelegram {
		status = "Опубликован в Telegram"
	} else if p.PublishedVK {
		status = "Опубликован в VK"
	} else if p.PublishError != "" {
		status = "Ошибка"
	}

	return status + " " + p.SourceChannel + ": " + contentPreview
}

// Age возвращает возраст поста в минутах
func (p *Post) Age() time.Duration {
	return time.Since(p.ParsedAt)
}

// IsFresh проверяет, является ли пост свежим (менее 24 часов)
func (p *Post) IsFresh() bool {
	return p.Age() < 24*time.Hour
}
