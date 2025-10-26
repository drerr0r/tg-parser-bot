package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// User модель пользователя
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse ответ на вход
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// JWTClaims claims для JWT токена
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
