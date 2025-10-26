package api

import (
	"net/http"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"go.uber.org/zap"
)

func SetupRoutes(ruleRepo *storage.RuleRepository, postRepo *storage.PostRepository, userRepo *storage.UserRepository, logRepo *storage.LogRepository, logger *zap.SugaredLogger, cfg *models.Config) http.Handler {
	handlers := NewHandlers(ruleRepo, postRepo, userRepo, logRepo, logger, cfg)
	mux := http.NewServeMux()

	// ========== ПУБЛИЧНЫЕ ENDPOINTS (ДО AuthMiddleware) ==========

	// Health check
	mux.HandleFunc("GET /health", handlers.HealthCheck)
	mux.HandleFunc("GET /api/health", handlers.HealthCheck)

	// Логин - публичный
	mux.HandleFunc("POST /api/auth/login", handlers.Login)

	// Frontend
	mux.HandleFunc("GET /", handlers.ServeFrontend)

	// ========== ЗАЩИЩЕННЫЕ ENDPOINTS (ПОСЛЕ AuthMiddleware) ==========

	// Auth (защищенный - требует токен)
	mux.HandleFunc("GET /api/auth/me", handlers.GetCurrentUser)

	// Rules API
	mux.HandleFunc("GET /api/rules", handlers.GetRules)
	mux.HandleFunc("POST /api/rules", handlers.CreateRule)
	mux.HandleFunc("PUT /api/rules/{id}", handlers.UpdateRule)
	mux.HandleFunc("DELETE /api/rules/{id}", handlers.DeleteRule)

	// Posts API
	mux.HandleFunc("GET /api/posts", handlers.GetPosts)

	// Stats
	mux.HandleFunc("GET /api/stats", handlers.GetStats)

	// Logs API
	mux.HandleFunc("GET /api/logs", handlers.GetLogs)

	// OPTIONS для CORS
	mux.HandleFunc("OPTIONS /api/{rest...}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Измените на ваш домен в продакшене
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})

	// ========== MIDDLEWARE ЦЕПОЧКА ==========
	// ВАЖНО: Правильный порядок!
	handler := CORSMiddleware(mux)
	handler = LoggingMiddleware(logger)(handler)
	handler = AuthMiddleware(cfg, logger)(handler)

	return handler
}
