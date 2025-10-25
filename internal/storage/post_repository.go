package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/jackc/pgx/v5"
)

// PostRepository репозиторий для работы с постами
type PostRepository struct {
	db *DB
}

// NewPostRepository создает новый репозиторий постов
func NewPostRepository(db *DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
        INSERT INTO posts (
            rule_id, message_id, source_channel, content, media_type,
            media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, parsed_at
    `

	err := r.db.Pool.QueryRow(ctx, query,
		post.RuleID,
		post.MessageID,
		post.SourceChannel,
		post.Content,
		post.MediaType,
		post.MediaURL,
		post.PostedAt,
		post.ParsedAt,
		post.PublishedTelegram, // новое поле
		post.PublishedVK,       // новое поле
		post.PublishError,
	).Scan(&post.ID, &post.ParsedAt)

	if err != nil {
		return fmt.Errorf("ошибка создания поста: %v", err)
	}

	return nil
}

// GetByID возвращает пост по ID
func (r *PostRepository) GetByID(ctx context.Context, id int64) (*models.Post, error) {
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type,
           media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts
    WHERE id = $1
`

	var post models.Post
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.RuleID,
		&post.MessageID,
		&post.SourceChannel,
		&post.Content,
		&post.MediaType,
		&post.MediaURL,
		&post.PostedAt,
		&post.ParsedAt,
		&post.PublishedTelegram,
		&post.PublishedVK,
		&post.PublishError,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения поста: %v", err)
	}

	return &post, nil
}

// GetByMessageID возвращает пост по ID сообщения и каналу
func (r *PostRepository) GetByMessageID(ctx context.Context, sourceChannel string, messageID int64) (*models.Post, error) {
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type,
           media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts
    WHERE source_channel = $1 AND message_id = $2
`

	var post models.Post
	err := r.db.Pool.QueryRow(ctx, query, sourceChannel, messageID).Scan(
		&post.ID,
		&post.RuleID,
		&post.MessageID,
		&post.SourceChannel,
		&post.Content,
		&post.MediaType,
		&post.MediaURL,
		&post.PostedAt,
		&post.ParsedAt,
		&post.PublishedTelegram,
		&post.PublishedVK,
		&post.PublishError,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения поста: %v", err)
	}

	return &post, nil
}

// MarkAsPublished помечает пост как опубликованный
func (r *PostRepository) MarkAsPublished(ctx context.Context, id int64) error {
	query := `UPDATE posts SET is_published = TRUE, publish_error = '' WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка отметки поста как опубликованного: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("пост с ID %d не найден", id)
	}

	return nil
}

// GetUnpublishedPosts возвращает неопубликованные посты
func (r *PostRepository) GetUnpublishedPosts(ctx context.Context, limit int) ([]*models.Post, error) {
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type,
           media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts
    WHERE (published_telegram = FALSE OR published_vk = FALSE) AND publish_error = ''
    ORDER BY parsed_at ASC
    LIMIT $1
`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса неопубликованных постов: %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.RuleID,
			&post.MessageID,
			&post.SourceChannel,
			&post.Content,
			&post.MediaType,
			&post.MediaURL,
			&post.PostedAt,
			&post.ParsedAt,
			&post.PublishedTelegram,
			&post.PublishedVK,
			&post.PublishError,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %v", err)
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

// GetPosts возвращает все посты с пагинацией
func (r *PostRepository) GetPosts(ctx context.Context, limit, offset int) ([]*models.Post, error) {
	// Если rule_id = 0, возвращаем все посты
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type,
           media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts
    ORDER BY parsed_at DESC
    LIMIT $1 OFFSET $2
`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса постов: %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.RuleID,
			&post.MessageID,
			&post.SourceChannel,
			&post.Content,
			&post.MediaType,
			&post.MediaURL,
			&post.PostedAt,
			&post.ParsedAt,
			&post.PublishedTelegram,
			&post.PublishedVK,
			&post.PublishError,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %v", err)
		}

		posts = append(posts, &post)
	}

	// Всегда возвращаем массив (даже пустой) вместо nil
	if posts == nil {
		posts = []*models.Post{}
	}

	return posts, nil
}

// GetPostsByRule возвращает посты по правилу
func (r *PostRepository) GetPostsByRule(ctx context.Context, ruleID int64, limit, offset int) ([]*models.Post, error) {
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type,
           media_url, posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts
    WHERE rule_id = $1
    ORDER BY parsed_at DESC
    LIMIT $2 OFFSET $3
`

	rows, err := r.db.Pool.Query(ctx, query, ruleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса постов по правилу: %v", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.RuleID,
			&post.MessageID,
			&post.SourceChannel,
			&post.Content,
			&post.MediaType,
			&post.MediaURL,
			&post.PostedAt,
			&post.ParsedAt,
			&post.PublishedTelegram,
			&post.PublishedVK,
			&post.PublishError,
		)

		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования поста: %v", err)
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

// GetStats возвращает статистику по постам
func (r *PostRepository) GetStats(ctx context.Context, ruleID int64, fromDate, toDate time.Time) (*PostStats, error) {
	query := `
        SELECT 
            COUNT(*) as total_posts,
            COUNT(CASE WHEN published_telegram = TRUE AND published_vk = TRUE THEN 1 END) as published_posts,
            COUNT(CASE WHEN (published_telegram = FALSE OR published_vk = FALSE) AND publish_error = '' THEN 1 END) as pending_posts,
            COUNT(CASE WHEN publish_error != '' THEN 1 END) as failed_posts
        FROM posts
        WHERE rule_id = $1 AND parsed_at BETWEEN $2 AND $3
    `

	var stats PostStats
	err := r.db.Pool.QueryRow(ctx, query, ruleID, fromDate, toDate).Scan(
		&stats.TotalPosts,
		&stats.PublishedPosts,
		&stats.PendingPosts,
		&stats.FailedPosts,
	)

	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики: %v", err)
	}

	return &stats, nil
}

// DeleteOldPosts удаляет старые посты
func (r *PostRepository) DeleteOldPosts(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM posts WHERE parsed_at < $1`

	result, err := r.db.Pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("ошибка удаления старых постов: %v", err)
	}

	return result.RowsAffected(), nil
}

// MarkAsPublishedTelegram помечает пост как опубликованный в Telegram
func (r *PostRepository) MarkAsPublishedTelegram(ctx context.Context, id int64) error {
	query := `UPDATE posts SET published_telegram = TRUE, publish_error = '' WHERE id = $1`
	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка отметки поста в Telegram: %v", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("пост с ID %d не найден", id)
	}
	return nil
}

// MarkAsPublishedVK помечает пост как опубликованный в VK
func (r *PostRepository) MarkAsPublishedVK(ctx context.Context, id int64) error {
	query := `UPDATE posts SET published_vk = TRUE, publish_error = '' WHERE id = $1`
	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка отметки поста в VK: %v", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("пост с ID %d не найден", id)
	}
	return nil
}

// MarkAsFailed помечает пост как неопубликованный с ошибкой
func (r *PostRepository) MarkAsFailed(ctx context.Context, id int64, errorMsg string) error {
	query := `UPDATE posts SET published_telegram = FALSE, published_vk = FALSE, publish_error = $1 WHERE id = $2`
	result, err := r.db.Pool.Exec(ctx, query, errorMsg, id)
	if err != nil {
		return fmt.Errorf("ошибка отметки поста как неудачного: %v", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("пост с ID %d не найден", id)
	}
	return nil
}

// PostStats статистика по постам
type PostStats struct {
	TotalPosts     int64 `json:"total_posts"`
	PublishedPosts int64 `json:"published_posts"`
	PendingPosts   int64 `json:"pending_posts"`
	FailedPosts    int64 `json:"failed_posts"`
}

// GetLastMessageID возвращает последний обработанный MessageID для канала
func (r *PostRepository) GetLastMessageID(ctx context.Context, sourceChannel string) (*models.Post, error) {
	query := `
    SELECT id, rule_id, message_id, source_channel, content, media_type, media_url, 
           posted_at, parsed_at, published_telegram, published_vk, publish_error
    FROM posts 
    WHERE source_channel = $1 
    ORDER BY message_id DESC 
    LIMIT 1
`

	var post models.Post
	err := r.db.Pool.QueryRow(ctx, query, sourceChannel).Scan(
		&post.ID, &post.RuleID, &post.MessageID, &post.SourceChannel, &post.Content,
		&post.MediaType, &post.MediaURL, &post.PostedAt, &post.ParsedAt,
		&post.PublishedTelegram,
		&post.PublishedVK, &post.PublishError,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения последнего сообщения: %v", err)
	}

	return &post, nil
}
