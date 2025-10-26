package storage

import (
	"context"
	"fmt"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/jackc/pgx/v5"
)

// UserRepository репозиторий для работы с пользователями
type UserRepository struct {
	db *DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByUsername возвращает пользователя по имени
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, email, role, is_active, created_at, updated_at
		FROM users
		WHERE username = $1 AND is_active = TRUE
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
	}

	return &user, nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, email, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
	}

	return &user, nil
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Email,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %v", err)
	}

	return nil
}
