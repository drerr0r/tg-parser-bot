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

// AuthMiddleware middleware для проверки JWT токена (net/http версия)
func AuthMiddleware(cfg *models.Config, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем публичные endpoints
			if r.URL.Path == "/api/auth/login" && r.Method == "POST" {
				next.ServeHTTP(w, r)
				return
			}

			// Пропускаем health check
			if r.URL.Path == "/health" || r.URL.Path == "/api/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, "Требуется авторизация", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен из заголовка
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
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)

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

// writeJSONError вспомогательная функция для отправки ошибок
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
