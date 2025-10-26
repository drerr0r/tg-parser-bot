package api

import (
	"net/http"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"go.uber.org/zap"
)

func SetupRoutes(ruleRepo *storage.RuleRepository, postRepo *storage.PostRepository, userRepo *storage.UserRepository, logger *zap.SugaredLogger, cfg *models.Config) http.Handler {
	handlers := NewHandlers(ruleRepo, postRepo, userRepo, logger, cfg)
	mux := http.NewServeMux()

	// Auth API
	mux.HandleFunc("POST /api/auth/login", handlers.Login)

	// Health check
	mux.HandleFunc("GET /health", handlers.HealthCheck)
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)

	// Rules API
	mux.HandleFunc("GET /api/rules", handlers.GetRules)
	mux.HandleFunc("POST /api/rules", handlers.CreateRule)
	mux.HandleFunc("PUT /api/rules/{id}", handlers.UpdateRule)
	mux.HandleFunc("DELETE /api/rules/{id}", handlers.DeleteRule)

	// Posts API
	mux.HandleFunc("GET /api/posts", handlers.GetPosts)

	// Stats
	mux.HandleFunc("GET /api/stats", handlers.GetStats)

	// User API
	mux.HandleFunc("GET /api/auth/me", handlers.GetCurrentUser)

	// Frontend - отдаем информационное сообщение
	mux.HandleFunc("GET /", handlers.ServeFrontend)

	// ПРАВИЛЬНЫЙ ПОРЯДОК: CORS -> Logging -> Auth
	handler := CORSMiddleware(mux)
	handler = LoggingMiddleware(logger)(handler)
	handler = AuthMiddleware(cfg, logger)(handler)

	return handler
}
