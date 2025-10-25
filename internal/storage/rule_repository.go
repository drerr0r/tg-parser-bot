package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/jackc/pgx/v5"
)

// RuleRepository репозиторий для работы с правилами парсинга
type RuleRepository struct {
	db *DB
}

// NewRuleRepository создает новый репозиторий правил
func NewRuleRepository(db *DB) *RuleRepository {
	return &RuleRepository{db: db}
}

// Create создает новое правило
func (r *RuleRepository) Create(ctx context.Context, rule *models.ParsingRule) error {
	query := `
        INSERT INTO parsing_rules (
            name, source_channel, keywords, exclude_words, media_types,
            min_text_length, max_text_length, text_replacements, add_prefix,
            add_suffix, target_platforms, check_interval, is_active, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        RETURNING id, created_at, updated_at
    `

	// Преобразуем текстовые замены в JSON
	replacementsJSON, err := json.Marshal(rule.TextReplacements)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга text_replacements: %v", err)
	}

	err = r.db.Pool.QueryRow(ctx, query,
		rule.Name,
		rule.SourceChannel,
		rule.Keywords,
		rule.ExcludeWords,
		rule.MediaTypes,
		rule.MinTextLength,
		rule.MaxTextLength,
		replacementsJSON,
		rule.AddPrefix,
		rule.AddSuffix,
		rule.TargetPlatforms,
		rule.CheckInterval,
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка создания правила: %v", err)
	}

	return nil
}

// GetByID возвращает правило по ID
func (r *RuleRepository) GetByID(ctx context.Context, id int64) (*models.ParsingRule, error) {
	query := `
		SELECT id, name, source_channel, keywords, exclude_words, media_types,
			   min_text_length, max_text_length, text_replacements, add_prefix,
			   add_suffix, target_platforms, check_interval, is_active, created_at, updated_at
		FROM parsing_rules
		WHERE id = $1
	`

	var rule models.ParsingRule
	var replacementsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.Name,
		&rule.SourceChannel,
		&rule.Keywords,
		&rule.ExcludeWords,
		&rule.MediaTypes,
		&rule.MinTextLength,
		&rule.MaxTextLength,
		&replacementsJSON,
		&rule.AddPrefix,
		&rule.AddSuffix,
		&rule.TargetPlatforms,
		&rule.CheckInterval,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения правила: %v", err)
	}

	// Парсим JSON замен текста
	if err := json.Unmarshal(replacementsJSON, &rule.TextReplacements); err != nil {
		return nil, fmt.Errorf("ошибка парсинга text_replacements: %v", err)
	}

	return &rule, nil
}

// GetActiveRules возвращает все активные правила
func (r *RuleRepository) GetActiveRules(ctx context.Context) ([]*models.ParsingRule, error) {
	query := `
        SELECT id, name, source_channel, keywords, exclude_words, media_types,
               min_text_length, max_text_length, text_replacements, add_prefix,
               add_suffix, target_platforms, check_interval, is_active, created_at, updated_at
        FROM parsing_rules 
        WHERE is_active = TRUE
        ORDER BY created_at DESC
    `

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса активных правил: %v", err)
	}
	defer rows.Close()

	var rules []*models.ParsingRule
	for rows.Next() {
		var rule models.ParsingRule
		var replacementsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.SourceChannel,
			&rule.Keywords,
			&rule.ExcludeWords,
			&rule.MediaTypes,
			&rule.MinTextLength,
			&rule.MaxTextLength,
			&replacementsJSON,
			&rule.AddPrefix,
			&rule.AddSuffix,
			&rule.TargetPlatforms,
			&rule.CheckInterval,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования правила: %v", err)
		}

		// Парсим JSON замен текста
		if err := json.Unmarshal(replacementsJSON, &rule.TextReplacements); err != nil {
			return nil, fmt.Errorf("ошибка парсинга text_replacements: %v", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetBySourceChannel возвращает правила для канала
func (r *RuleRepository) GetBySourceChannel(ctx context.Context, sourceChannel string) ([]*models.ParsingRule, error) {
	query := `
		SELECT id, name, source_channel, keywords, exclude_words, media_types,
			   min_text_length, max_text_length, text_replacements, add_prefix,
			   add_suffix, target_platforms, check_interval, is_active, created_at, updated_at
		FROM parsing_rules
		WHERE source_channel = $1 AND is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, sourceChannel)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса правил по каналу: %v", err)
	}
	defer rows.Close()

	var rules []*models.ParsingRule
	for rows.Next() {
		var rule models.ParsingRule
		var replacementsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.SourceChannel,
			&rule.Keywords,
			&rule.ExcludeWords,
			&rule.MediaTypes,
			&rule.MinTextLength,
			&rule.MaxTextLength,
			&replacementsJSON,
			&rule.AddPrefix,
			&rule.AddSuffix,
			&rule.TargetPlatforms,
			&rule.CheckInterval,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования правила: %v", err)
		}

		// Парсим JSON замен текста
		if err := json.Unmarshal(replacementsJSON, &rule.TextReplacements); err != nil {
			return nil, fmt.Errorf("ошибка парсинга text_replacements: %v", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

// Update обновляет правило
func (r *RuleRepository) Update(ctx context.Context, rule *models.ParsingRule) error {
	query := `
		UPDATE parsing_rules 
		SET name = $1, source_channel = $2, keywords = $3, exclude_words = $4,
			media_types = $5, min_text_length = $6, max_text_length = $7,
			text_replacements = $8, add_prefix = $9, add_suffix = $10,
			target_platforms = $11, check_interval = $12, is_active = $13, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
		RETURNING updated_at
	`

	replacementsJSON, err := json.Marshal(rule.TextReplacements)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга text_replacements: %v", err)
	}

	err = r.db.Pool.QueryRow(ctx, query,
		rule.Name,
		rule.SourceChannel,
		rule.Keywords,
		rule.ExcludeWords,
		rule.MediaTypes,
		rule.MinTextLength,
		rule.MaxTextLength,
		replacementsJSON,
		rule.AddPrefix,
		rule.AddSuffix,
		rule.TargetPlatforms,
		rule.CheckInterval,
		rule.IsActive,
		rule.ID,
	).Scan(&rule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка обновления правила: %v", err)
	}

	return nil
}

// Delete удаляет правило
func (r *RuleRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM parsing_rules WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления правила: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("правило с ID %d не найдено", id)
	}

	return nil
}

// List возвращает все правила с пагинацией
func (r *RuleRepository) List(ctx context.Context, limit, offset int) ([]*models.ParsingRule, error) {
	query := `
		SELECT id, name, source_channel, keywords, exclude_words, media_types,
			   min_text_length, max_text_length, text_replacements, add_prefix,
			   add_suffix, target_platforms, is_active, created_at, updated_at
		FROM parsing_rules
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса списка правил: %v", err)
	}
	defer rows.Close()

	var rules []*models.ParsingRule
	for rows.Next() {
		var rule models.ParsingRule
		var replacementsJSON []byte

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.SourceChannel,
			&rule.Keywords,
			&rule.ExcludeWords,
			&rule.MediaTypes,
			&rule.MinTextLength,
			&rule.MaxTextLength,
			&replacementsJSON,
			&rule.AddPrefix,
			&rule.AddSuffix,
			&rule.TargetPlatforms,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования правила: %v", err)
		}

		if err := json.Unmarshal(replacementsJSON, &rule.TextReplacements); err != nil {
			return nil, fmt.Errorf("ошибка парсинга text_replacements: %v", err)
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}
