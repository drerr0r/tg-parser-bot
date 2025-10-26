package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drerr0r/tgparserbot/internal/api"
	"github.com/drerr0r/tgparserbot/internal/config"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"github.com/drerr0r/tgparserbot/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализируем логгер
	if err := logger.Init(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.FilePath); err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()

	logger.Sugar().Info("🚀 Запуск Web UI сервера...")

	// Валидируем конфигурацию
	if err := config.Validate(cfg); err != nil {
		logger.Sugar().Fatalf("Ошибка валидации конфигурации: %v", err)
	}

	// Инициализируем хранилище
	db, err := storage.New(cfg.Database)
	if err != nil {
		logger.Sugar().Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	logger.Sugar().Info("✅ Успешное подключение к БД")

	// УМНАЯ ПРОВЕРКА И ПРИМЕНЕНИЕ МИГРАЦИЙ
	migrationPaths := []string{
		"./migrations",  // относительный путь
		"migrations",    // текущая директория
		"../migrations", // на уровень выше
	}

	var migrationApplied bool
	var migrationErr error

	for _, path := range migrationPaths {
		logger.Sugar().Infof("🔍 Проверяем миграции в: %s", path)

		if _, err := os.Stat(path); err == nil {
			// Папка существует, пробуем применить миграции
			if err := db.RunMigrations(path); err == nil {
				logger.Sugar().Infof("✅ Миграции успешно применены из: %s", path)
				migrationApplied = true
				migrationErr = nil
				break
			} else {
				migrationErr = err
				logger.Sugar().Debugf("❌ Ошибка в пути %s: %v", path, err)
			}
		}
	}

	if !migrationApplied {
		if migrationErr != nil {
			logger.Sugar().Errorf("❌ Критическая ошибка: не удалось применить миграции: %v", migrationErr)
			logger.Sugar().Fatal("❌ Невозможно запустить приложение без миграций БД")
		} else {
			logger.Sugar().Warn("⚠️ Папка миграций не найден, проверяем состояние БД...")
		}
	}

	// Финальная проверка состояния БД
	if !db.CheckMigrationsApplied() {
		logger.Sugar().Fatal("❌ Критическая ошибка: БД не соответствует требуемой схеме. Проверьте миграции.")
	} else {
		logger.Sugar().Info("✅ Проверка схемы БД пройдена успешно")
	}

	// Инициализируем репозитории
	ruleRepo := storage.NewRuleRepository(db)
	postRepo := storage.NewPostRepository(db)
	userRepo := storage.NewUserRepository(db)

	// ДОБАВЬТЕ ЭТУ СТРОКУ - создаем репозиторий логов
	logRepo := storage.NewLogRepository("logs/app.log")

	logger.Sugar().Info("✅ Репозитории инициализированы")

	// Настраиваем HTTP сервер
	// ОБНОВИТЕ ЭТУ СТРОКУ - добавьте logRepo как 4-й аргумент
	handler := api.SetupRoutes(ruleRepo, postRepo, userRepo, logRepo, logger.Sugar(), cfg)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		logger.Sugar().Infof("🌐 Web сервер запущен на http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		logger.Sugar().Info("📋 Доступные endpoints:")
		logger.Sugar().Info("   GET /health - Проверка здоровья")
		logger.Sugar().Info("   GET /api/rules - Список правил")
		logger.Sugar().Info("   GET /api/posts - Список постов")
		logger.Sugar().Info("   GET /api/stats - Статистика")
		logger.Sugar().Info("   GET /api/logs - Просмотр логов") // ДОБАВЬТЕ ЭТУ СТРОКУ

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Fatalf("❌ Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигналы завершения
	waitForShutdown(server, logger.Sugar())
}

func waitForShutdown(server *http.Server, logger *zap.SugaredLogger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("🛑 Получен сигнал завершения. Остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("❌ Ошибка graceful shutdown: %v", err)
	} else {
		logger.Info("✅ Сервер успешно остановлен")
	}
}
