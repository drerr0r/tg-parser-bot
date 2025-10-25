package parser

import (
	"context"
	"fmt"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
)

// MessageProcessor обработчик сообщений
type MessageProcessor struct {
	ruleRepo   *storage.RuleRepository
	postRepo   *storage.PostRepository
	ruleEngine *RuleEngine
}

// NewMessageProcessor создает новый обработчик сообщений
func NewMessageProcessor(
	ruleRepo *storage.RuleRepository,
	postRepo *storage.PostRepository,
) *MessageProcessor {
	return &MessageProcessor{
		ruleRepo:   ruleRepo,
		postRepo:   postRepo,
		ruleEngine: NewRuleEngine(),
	}
}

// Process обрабатывает входящее сообщение
func (p *MessageProcessor) Process(ctx context.Context, message *ParsedMessage) error {
	// Получаем правила для канала
	rules, err := p.ruleRepo.GetBySourceChannel(ctx, message.SourceChannel)
	if err != nil {
		return fmt.Errorf("ошибка получения правил: %v", err)
	}

	// Применяем правила
	matchingRules := p.ruleEngine.ProcessMessage(message, rules)

	// Обрабатываем для каждого подходящего правила
	for _, rule := range matchingRules {
		if err := p.processForRule(ctx, message, rule); err != nil {
			return fmt.Errorf("ошибка обработки для правила %s: %v", rule.Name, err)
		}
	}

	return nil
}

// processForRule обрабатывает сообщение для конкретного правила
func (p *MessageProcessor) processForRule(ctx context.Context, message *ParsedMessage, rule *models.ParsingRule) error {
	// Проверяем, не обрабатывали ли мы уже это сообщение для этого правила
	existingPost, err := p.postRepo.GetByMessageID(ctx, message.SourceChannel, message.ID)
	if err != nil {
		return fmt.Errorf("ошибка проверки существующего поста: %v", err)
	}

	if existingPost != nil {
		return nil // Уже обработано
	}

	// Применяем трансформации
	transformedContent := p.ruleEngine.ApplyRule(message.Content, rule)

	// Создаем пост
	post := models.NewPost(rule.ID, message.ID, message.SourceChannel, transformedContent, message.MediaType)
	post.MediaURL = message.MediaURL
	post.PostedAt = message.Date

	if err := post.Validate(); err != nil {
		return fmt.Errorf("ошибка валидации поста: %v", err)
	}

	// Сохраняем в БД
	if err := p.postRepo.Create(ctx, post); err != nil {
		return fmt.Errorf("ошибка сохранения поста: %v", err)
	}

	return nil
}
