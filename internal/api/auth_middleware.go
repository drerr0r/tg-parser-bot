package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Кастомные типы для ключей контекста (исправление SA1029)
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
	roleKey     contextKey = "role"
)

func AuthMiddleware(cfg *models.Config, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем OPTIONS запросы
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Пропускаем публичные endpoints
			publicPaths := map[string]bool{
				"/health":         true,
				"/api/health":     true,
				"/api/auth/login": true,
				"/":               true,
			}

			if publicPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Для защищенных endpoints проверяем авторизацию
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			// ... остальная логика проверки токена
			tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
			if tokenString == "" {
				writeJSONError(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			// Парсим и валидируем токен
			claims := &models.JWTClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Auth.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				writeJSONError(w, "Неверный или просроченный токен", http.StatusUnauthorized)
				return
			}

			// Сохраняем информацию о пользователе в контекст
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, usernameKey, claims.Username)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GenerateJWTToken генерирует JWT токен для пользователя
func GenerateJWTToken(user *models.User, jwtSecret string, jwtDuration int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(jwtDuration) * time.Hour)

	claims := &models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "tg-parser-bot",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// Вспомогательные функции для получения значений из контекста
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}

// writeJSONError вспомогательная функция для отправки ошибок
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
