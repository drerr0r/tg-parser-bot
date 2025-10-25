package api

import (
	"net/http"

	"github.com/drerr0r/tgparserbot/internal/storage"
	"go.uber.org/zap"
)

func SetupRoutes(ruleRepo *storage.RuleRepository, postRepo *storage.PostRepository, logger *zap.SugaredLogger) http.Handler {
	handlers := NewHandlers(ruleRepo, postRepo, logger)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", handlers.HealthCheck)
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)

	// Rules API
	mux.HandleFunc("GET /api/rules", handlers.GetRules)
	mux.HandleFunc("POST /api/rules", handlers.CreateRule)
	mux.HandleFunc("PUT /api/rules/{id}", handlers.UpdateRule)    // Исправлено на path parameter
	mux.HandleFunc("DELETE /api/rules/{id}", handlers.DeleteRule) // Исправлено на path parameter

	// Posts API
	mux.HandleFunc("GET /api/posts", handlers.GetPosts)

	// Stats
	mux.HandleFunc("GET /api/stats", handlers.GetStats)

	// Frontend - отдаем информационное сообщение
	mux.HandleFunc("GET /", handlers.ServeFrontend)

	// Оборачиваем в middleware (CORSMiddleware уже есть в middleware.go)
	handler := LoggingMiddleware(logger)(mux)
	handler = CORSMiddleware(handler) // Используем существующий middleware

	return handler
}
