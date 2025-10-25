package models

import (
	"errors"
	"strings"
	"time"
)

// NewParsingRule создает новое правило с настройками по умолчанию
func NewParsingRule(name, sourceChannel string) *ParsingRule {
	return &ParsingRule{
		Name:             name,
		SourceChannel:    sourceChannel,
		Keywords:         []string{},
		ExcludeWords:     []string{},
		MediaTypes:       []MediaType{MediaText, MediaPhoto},
		MinTextLength:    0,
		MaxTextLength:    0,
		TextReplacements: make(map[string]string),
		AddPrefix:        "",
		AddSuffix:        "",
		TargetPlatforms:  []PlatformType{PlatformTelegram},
		CheckInterval:    2,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// MatchesKeywords проверяет, соответствует ли текст ключевым словам
func (r *ParsingRule) MatchesKeywords(text string) bool {
	if len(r.Keywords) == 0 {
		return true
	}

	text = strings.ToLower(text)
	for _, keyword := range r.Keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// ContainsExcludedWords проверяет, содержит ли текст слова-исключения
func (r *ParsingRule) ContainsExcludedWords(text string) bool {
	if len(r.ExcludeWords) == 0 {
		return false
	}

	text = strings.ToLower(text)
	for _, word := range r.ExcludeWords {
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// SupportsMediaType проверяет, поддерживается ли тип медиа
func (r *ParsingRule) SupportsMediaType(mediaType MediaType) bool {
	if len(r.MediaTypes) == 0 {
		return true
	}

	for _, mt := range r.MediaTypes {
		if mt == mediaType {
			return true
		}
	}
	return false
}

// TextLengthValid проверяет длину текста
func (r *ParsingRule) TextLengthValid(text string) bool {
	length := len(text)

	if r.MinTextLength > 0 && length < r.MinTextLength {
		return false
	}
	if r.MaxTextLength > 0 && length > r.MaxTextLength {
		return false
	}
	return true
}

// ApplyTransformations применяет трансформации к тексту
func (r *ParsingRule) ApplyTransformations(text string) string {
	result := text

	for old, new := range r.TextReplacements {
		result = strings.ReplaceAll(result, old, new)
	}

	if r.AddPrefix != "" {
		result = r.AddPrefix + result
	}
	if r.AddSuffix != "" {
		result = result + r.AddSuffix
	}

	return result
}

// SupportsPlatform проверяет, поддерживается ли платформа для публикации
func (r *ParsingRule) SupportsPlatform(platform PlatformType) bool {
	for _, p := range r.TargetPlatforms {
		if p == platform {
			return true
		}
	}
	return false
}

// Validate проверяет валидность правила
func (r *ParsingRule) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.SourceChannel == "" {
		return errors.New("source channel is required")
	}
	if len(r.TargetPlatforms) == 0 {
		return errors.New("at least one target platform is required")
	}
	return nil
}
