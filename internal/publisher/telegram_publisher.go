package publisher

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/drerr0r/tgparserbot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TelegramPublisher публикатор в Telegram
type TelegramPublisher struct {
	bot    *tgbotapi.BotAPI
	logger *zap.SugaredLogger
	cfg    *models.TelegramConfig
}

// NewTelegramPublisher создает новый публикатор для Telegram
func NewTelegramPublisher(cfg *models.TelegramConfig, logger *zap.SugaredLogger) (*TelegramPublisher, error) {
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("bot_token не указан в конфигурации")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота Telegram: %v", err)
	}

	logger.Infof("Авторизован в Telegram как %s", bot.Self.UserName)

	return &TelegramPublisher{
		bot:    bot,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// Publish публикует пост в Telegram канал
func (p *TelegramPublisher) Publish(ctx context.Context, post *models.Post, targetChannel string) error {
	p.logger.Infof("Публикация поста %d в канал %s", post.ID, targetChannel)

	// Всегда используем target_channel из конфига, игнорируя переданный параметр
	if p.cfg.TargetChannel == "" {
		return fmt.Errorf("target_channel не указан в конфигурации")
	}

	// Парсим ChatID из конфига
	chatID, err := strconv.ParseInt(p.cfg.TargetChannel, 10, 64)
	if err != nil {
		return fmt.Errorf("ошибка парсинга ChatID из конфига %s: %v", p.cfg.TargetChannel, err)
	}

	// Подготавливаем контент
	content := p.prepareContent(post)

	// Если есть медиа, публикуем с медиа
	if post.MediaURL != "" {
		return p.publishWithMedia(post, chatID, content)
	}

	// Публикуем текстовое сообщение
	return p.publishText(chatID, content)
}

// publishWithMedia публикует пост с медиа
func (p *TelegramPublisher) publishWithMedia(post *models.Post, chatID int64, content string) error {
	switch post.MediaType {
	case models.MediaPhoto:
		return p.publishPhoto(post, chatID, content)
	case models.MediaVideo:
		return p.publishVideo(post, chatID, content)
	case models.MediaDocument:
		return p.publishDocument(post, chatID, content)
	default:
		return p.publishText(chatID, content)
	}
}

// publishPhoto публикует фото
func (p *TelegramPublisher) publishPhoto(post *models.Post, chatID int64, caption string) error {
	// Создаем конфиг для фото с URL
	photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(post.MediaURL))
	photo.Caption = caption
	photo.ParseMode = "HTML"

	_, err := p.bot.Send(photo)
	if err != nil {
		return fmt.Errorf("ошибка публикации фото: %v", err)
	}
	return nil
}

// publishVideo публикует видео
func (p *TelegramPublisher) publishVideo(post *models.Post, chatID int64, caption string) error {
	// Создаем конфиг для видео с URL
	video := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(post.MediaURL))
	video.Caption = caption
	video.ParseMode = "HTML"

	_, err := p.bot.Send(video)
	if err != nil {
		return fmt.Errorf("ошибка публикации видео: %v", err)
	}
	return nil
}

// publishDocument публикует документ
func (p *TelegramPublisher) publishDocument(post *models.Post, chatID int64, caption string) error {
	// Создаем конфиг для документа с URL
	document := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(post.MediaURL))
	document.Caption = caption
	document.ParseMode = "HTML"

	_, err := p.bot.Send(document)
	if err != nil {
		return fmt.Errorf("ошибка публикации документа: %v", err)
	}
	return nil
}

// publishText публикует текстовое сообщение
func (p *TelegramPublisher) publishText(chatID int64, content string) error {
	msg := tgbotapi.NewMessage(chatID, content)
	msg.ParseMode = "HTML"

	_, err := p.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("ошибка публикации текста: %v", err)
	}
	return nil
}

// prepareContent подготавливает контент для публикации
func (p *TelegramPublisher) prepareContent(post *models.Post) string {
	var content strings.Builder

	// Экранируем HTML-символы
	escapedContent := strings.ReplaceAll(post.Content, "&", "&amp;")
	escapedContent = strings.ReplaceAll(escapedContent, "<", "&lt;")
	escapedContent = strings.ReplaceAll(escapedContent, ">", "&gt;")

	content.WriteString(escapedContent)

	// Добавляем информацию об источнике
	if post.SourceChannel != "" {
		content.WriteString(fmt.Sprintf("\n\n📎 <i>Источник: %s</i>", post.SourceChannel))
	}

	return content.String()
}

// TestConnection проверяет подключение к Telegram
func (p *TelegramPublisher) TestConnection(ctx context.Context) error {
	_, err := p.bot.GetMe()
	if err != nil {
		return fmt.Errorf("ошибка проверки подключения к Telegram: %v", err)
	}
	return nil
}
