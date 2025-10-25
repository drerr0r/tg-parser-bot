package publisher

import (
	"context"
	"fmt"
	"strings"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"go.uber.org/zap"
)

// MultiPublisher управляет публикацией в multiple платформы
type MultiPublisher struct {
	tgPublisher *TelegramPublisher
	vkPublisher *VKPublisher
	postRepo    *storage.PostRepository
	logger      *zap.SugaredLogger
	tgChannelID string // ChatID для Telegram канала
}

// NewMultiPublisher создает новый мульти-публикатор
func NewMultiPublisher(
	tgPublisher *TelegramPublisher,
	vkPublisher *VKPublisher,
	postRepo *storage.PostRepository,
	tgChannelID string,
	logger *zap.SugaredLogger,
) *MultiPublisher {
	return &MultiPublisher{
		tgPublisher: tgPublisher,
		vkPublisher: vkPublisher,
		postRepo:    postRepo,
		tgChannelID: tgChannelID,
		logger:      logger,
	}
}

// containsPlatform проверяет наличие платформы в массиве
func containsPlatform(platforms []models.PlatformType, platform string) bool {
	for _, p := range platforms {
		if string(p) == platform {
			return true
		}
	}
	return false
}
func (p *MultiPublisher) Publish(ctx context.Context, post *models.Post, rule *models.ParsingRule) error {
	p.logger.Infof("🔄 Начало публикации поста %d на платформы: %v", post.ID, rule.TargetPlatforms)

	var errors []string
	publishedPlatforms := []string{}

	// Публикация в Telegram
	if p.tgPublisher != nil && containsPlatform(rule.TargetPlatforms, "telegram") {
		p.logger.Infof("📤 Публикация поста %d на telegram", post.ID)
		if err := p.tgPublisher.Publish(ctx, post, p.tgChannelID); err != nil {
			p.logger.Errorf("❌ Ошибка публикации в Telegram: %v", err)
			errors = append(errors, fmt.Sprintf("Telegram: %v", err))
		} else {
			// Обновляем статус в БД
			if err := p.postRepo.MarkAsPublishedTelegram(ctx, post.ID); err != nil {
				p.logger.Errorf("❌ Ошибка обновления статуса Telegram: %v", err)
			} else {
				publishedPlatforms = append(publishedPlatforms, "telegram")
				p.logger.Infof("✅ Успешная публикация поста %d на telegram", post.ID)
			}
		}
	}

	// Публикация в VK
	if p.vkPublisher != nil && containsPlatform(rule.TargetPlatforms, "vk") {
		p.logger.Infof("📤 Публикация поста %d на vk", post.ID)
		if err := p.vkPublisher.Publish(ctx, post); err != nil {
			p.logger.Errorf("❌ Ошибка публикации в VK: %v", err)
			errors = append(errors, fmt.Sprintf("VK: %v", err))
		} else {
			// Обновляем статус в БД
			if err := p.postRepo.MarkAsPublishedVK(ctx, post.ID); err != nil {
				p.logger.Errorf("❌ Ошибка обновления статуса VK: %v", err)
			} else {
				publishedPlatforms = append(publishedPlatforms, "vk")
				p.logger.Infof("✅ Успешная публикация поста %d на vk", post.ID)
			}
		}
	}

	if len(errors) > 0 {
		errorMsg := strings.Join(errors, "; ")
		if err := p.postRepo.MarkAsFailed(ctx, post.ID, errorMsg); err != nil {
			p.logger.Errorf("❌ Ошибка сохранения ошибки публикации: %v", err)
		}
		return fmt.Errorf("ошибки публикации: %s", errorMsg)
	}

	p.logger.Infof("🎉 Пост %d успешно опубликован на платформы: %v", post.ID, publishedPlatforms)
	return nil
}

// ProcessUnpublished обрабатывает неопубликованные посты
func (p *MultiPublisher) ProcessUnpublished(ctx context.Context) error {
	p.logger.Info("Обработка неопубликованных постов...")

	// Получаем неопубликованные посты
	posts, err := p.postRepo.GetUnpublishedPosts(ctx, 10) // Ограничиваем 10 постами за раз
	if err != nil {
		return fmt.Errorf("ошибка получения неопубликованных постов: %v", err)
	}

	if len(posts) == 0 {
		p.logger.Info("Нет неопубликованных постов для обработки")
		return nil
	}

	p.logger.Infof("Найдено %d неопубликованных постов", len(posts))

	// Обрабатываем каждый пост
	for _, post := range posts {
		// Получаем правило для поста
		// TODO: Реализовать получение правила по post.RuleID

		// Временная заглушка - создаем mock правило
		rule := &models.ParsingRule{
			TargetPlatforms: []models.PlatformType{models.PlatformTelegram, models.PlatformVK},
		}

		if err := p.Publish(ctx, post, rule); err != nil {
			p.logger.Errorf("Ошибка публикации поста %d: %v", post.ID, err)
			continue
		}
	}

	return nil
}

// TestConnections проверяет подключения ко всем платформам
func (p *MultiPublisher) TestConnections(ctx context.Context) error {
	p.logger.Info("Проверка подключений к платформам...")

	if p.tgPublisher != nil {
		if err := p.tgPublisher.TestConnection(ctx); err != nil {
			return fmt.Errorf("ошибка подключения Telegram: %v", err)
		}
		p.logger.Info("✓ Подключение к Telegram OK")
	}

	if p.vkPublisher != nil {
		if err := p.vkPublisher.TestConnection(ctx); err != nil {
			return fmt.Errorf("ошибка подключения VK: %v", err)
		}
		p.logger.Info("✓ Подключение к VK OK")
	}

	p.logger.Info("Все подключения работают")
	return nil
}
