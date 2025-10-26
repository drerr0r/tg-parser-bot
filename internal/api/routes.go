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

	// Явно обрабатываем OPTIONS для ВСЕХ API endpoints
	mux.HandleFunc("OPTIONS /api/{rest...}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})

	// Auth API
	mux.HandleFunc("POST /api/auth/login", handlers.Login)
	mux.HandleFunc("GET /api/auth/me", handlers.GetCurrentUser)

	// Health check (должны быть ДО auth middleware)
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

	// Frontend
	mux.HandleFunc("GET /", handlers.ServeFrontend)

	// ВАЖНО: CORS должен быть ПЕРВЫМ в цепочке
	handler := CORSMiddleware(mux)
	handler = LoggingMiddleware(logger)(handler)
	handler = AuthMiddleware(cfg, logger)(handler)

	// Logs API
	mux.HandleFunc("GET /api/logs", handlers.GetLogs)

	return handler
}
