package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drerr0r/tgparserbot/internal/config"
	"github.com/drerr0r/tgparserbot/internal/parser"
	"github.com/drerr0r/tgparserbot/internal/publisher"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"github.com/drerr0r/tgparserbot/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация логгера
	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	defer logger.Sync()

	sugar.Info("🚀 Запуск Telegram парсера...")

	// Подключение к БД
	db, err := storage.New(cfg.Database)
	if err != nil {
		sugar.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	sugar.Info("✅ Успешное подключение к БД")

	// Инициализация репозиториев
	ruleRepo := storage.NewRuleRepository(db)
	postRepo := storage.NewPostRepository(db)

	// Инициализация MTProto клиента
	mtprotoClient := parser.NewMTProtoClient(
		cfg.Telegram.APIID,
		cfg.Telegram.APIHash,
		cfg.Telegram.Phone,
		cfg.Telegram.SessionFile,
		sugar,
	)

	// Инициализация паблишеров
	var tgPublisher *publisher.TelegramPublisher
	var vkPublisher *publisher.VKPublisher

	if cfg.Telegram.BotToken != "" {
		tgPublisher, err = publisher.NewTelegramPublisher(&cfg.Telegram, sugar) // ПЕРЕДАЕМ ВЕСЬ КОНФИГ
		if err != nil {
			sugar.Errorf("❌ Ошибка инициализации Telegram публикатора: %v", err)
		} else {
			sugar.Info("✅ Telegram публикатор инициализирован")
		}
	}
	if cfg.VK.AccessToken != "" {
		vkPublisher, err = publisher.NewVKPublisher(cfg.VK.AccessToken, cfg.VK.GroupID, sugar)
		if err != nil {
			sugar.Errorf("❌ Ошибка инициализации VK публикатора: %v", err)
		} else {
			sugar.Info("✅ VK публикатор инициализирован")
		}
	}

	// Создаем MultiPublisher
	multiPublisher := publisher.NewMultiPublisher(
		tgPublisher,
		vkPublisher,
		postRepo,
		cfg.Telegram.TargetChannel,
		sugar,
	)

	// Инициализация парсера
	telegramParser := parser.NewTelegramParser(
		db,
		ruleRepo,
		postRepo,
		multiPublisher,
		mtprotoClient,
		sugar,
	)

	// Обработка сигналов завершения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ЗАПУСКАЕМ ФОНОВЫЙ ПРОЦЕСС ПУБЛИКАЦИИ (после создания ctx)
	go func() {
		sugar.Info("🚀 Запуск фонового процесса публикации...")

		ticker := time.NewTicker(30 * time.Second) // Проверяем каждые 30 секунд
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := publishUnpublishedPosts(ctx, postRepo, ruleRepo, multiPublisher); err != nil {
					sugar.Errorf("❌ Ошибка публикации постов: %v", err)
				}
			case <-ctx.Done():
				sugar.Info("🛑 Остановка процесса публикации")
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		sugar.Infof("📞 Получен сигнал: %v", sig)
		telegramParser.Stop()
		cancel()
	}()

	// Запуск парсера с обработкой ошибок и перезапуском
	for {
		sugar.Info("🔄 Запуск парсера...")

		if err := telegramParser.Start(ctx); err != nil {
			sugar.Errorf("❌ Ошибка запуска парсера: %v", err)
			sugar.Info("🔄 Перезапуск через 30 секунд...")

			select {
			case <-time.After(30 * time.Second):
				continue // Перезапускаем
			case <-ctx.Done():
				sugar.Info("👋 Завершение работы по сигналу")
				return
			}
		} else {
			sugar.Info("✅ Парсер запущен. Ожидание сообщений...")
		}

		// Ожидание завершения
		<-ctx.Done()
		sugar.Info("👋 Завершение работы парсера")
		return
	}
}

// publishUnpublishedPosts публикует неопубликованные посты
func publishUnpublishedPosts(ctx context.Context, postRepo *storage.PostRepository, ruleRepo *storage.RuleRepository, publisher *publisher.MultiPublisher) error {
	// Получаем неопубликованные посты
	posts, err := postRepo.GetUnpublishedPosts(ctx, 10)
	if err != nil {
		return fmt.Errorf("ошибка получения постов: %v", err)
	}

	if len(posts) > 0 {
		logger.Sugar().Infof("📤 Найдено %d постов для публикации", len(posts))
	}

	for _, post := range posts {
		// Получаем правило для поста
		rule, err := ruleRepo.GetByID(ctx, post.RuleID)
		if err != nil {
			logger.Sugar().Errorf("❌ Ошибка получения правила для поста %d: %v", post.ID, err)
			continue
		}

		if rule == nil {
			logger.Sugar().Errorf("❌ Правило не найдено для поста %d", post.ID)
			continue
		}

		logger.Sugar().Infof("🔄 Публикую пост ID %d: %s", post.ID, post.GetSummary())

		// Публикуем пост (теперь с правильным количеством аргументов)
		if err := publisher.Publish(ctx, post, rule); err != nil {
			logger.Sugar().Errorf("❌ Ошибка публикации поста %d: %v", post.ID, err)

			// Отмечаем пост как неудачный
			if markErr := postRepo.MarkAsFailed(ctx, post.ID, err.Error()); markErr != nil {
				logger.Sugar().Errorf("❌ Ошибка отметки поста как неудачного: %v", markErr)
			}
		} else {
			logger.Sugar().Infof("✅ Пост %d успешно опубликован", post.ID)
		}
	}

	return nil
}
