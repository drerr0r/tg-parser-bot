package parser

import (
	"github.com/drerr0r/tgparserbot/internal/models"
)

// RuleEngine движок для применения правил
type RuleEngine struct{}

// NewRuleEngine создает новый движок правил
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// ProcessMessage применяет все правила к сообщению
func (e *RuleEngine) ProcessMessage(message *ParsedMessage, rules []*models.ParsingRule) []*models.ParsingRule {
	var matchingRules []*models.ParsingRule

	for _, rule := range rules {
		if e.matchesRule(message, rule) {
			matchingRules = append(matchingRules, rule)
		}
	}

	return matchingRules
}

// matchesRule проверяет соответствие сообщения правилу
func (e *RuleEngine) matchesRule(message *ParsedMessage, rule *models.ParsingRule) bool {
	// Проверка ключевых слов
	if !rule.MatchesKeywords(message.Content) {
		return false
	}

	// Проверка слов-исключений
	if rule.ContainsExcludedWords(message.Content) {
		return false
	}

	// Проверка типа медиа
	if !rule.SupportsMediaType(message.MediaType) {
		return false
	}

	// Проверка длины текста
	if !rule.TextLengthValid(message.Content) {
		return false
	}

	return true
}

// ApplyRule применяет трансформации правила к контенту
func (e *RuleEngine) ApplyRule(content string, rule *models.ParsingRule) string {
	return rule.ApplyTransformations(content)
}
