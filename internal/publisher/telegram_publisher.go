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
}

// NewTelegramPublisher создает новый публикатор для Telegram
func NewTelegramPublisher(botToken string, logger *zap.SugaredLogger) (*TelegramPublisher, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания бота Telegram: %v", err)
	}

	logger.Infof("Авторизован в Telegram как %s", bot.Self.UserName)

	return &TelegramPublisher{
		bot:    bot,
		logger: logger,
	}, nil
}

// Publish публикует пост в Telegram канал
func (p *TelegramPublisher) Publish(ctx context.Context, post *models.Post, targetChannel string) error {
	p.logger.Infof("Публикация поста %d в канал %s", post.ID, targetChannel)

	// Конвертируем channel name в ChatID
	chatID, err := p.getChatID(targetChannel)
	if err != nil {
		return fmt.Errorf("ошибка получения ChatID для %s: %v", targetChannel, err)
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

// getChatID конвертирует channel name или ID в int64
func (p *TelegramPublisher) getChatID(target string) (int64, error) {
	// Если target начинается с @, это username канала
	if strings.HasPrefix(target, "@") {
		// Для каналов по username нужно отправить сообщение и получить chat_id из ответа
		// Или использовать метод getUpdates чтобы найти канал
		chatID, err := p.findChatIDByUsername(target)
		if err != nil {
			return 0, fmt.Errorf("не удалось найти chat_id для %s: %v", target, err)
		}
		return chatID, nil
	}

	// Если target - числовой ID
	chatID, err := strconv.ParseInt(target, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("неверный формат ChatID: %s", target)
	}

	return chatID, nil
}

// findChatIDByUsername находит chat_id канала по username
func (p *TelegramPublisher) findChatIDByUsername(username string) (int64, error) {
	// Метод 1: Попробуем отправить тестовое сообщение и получить chat_id из ответа
	msg := tgbotapi.NewMessageToChannel(username, "test")
	sentMsg, err := p.bot.Send(msg)
	if err != nil {
		// Метод 2: Используем getUpdates чтобы найти канал
		return p.getChatIDFromUpdates(username)
	}
	return sentMsg.Chat.ID, nil
}

// getChatIDFromUpdates ищет chat_id в списке обновлений
func (p *TelegramPublisher) getChatIDFromUpdates(username string) (int64, error) {
	updates, err := p.bot.GetUpdates(tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   100,
		Timeout: 10,
	})
	if err != nil {
		return 0, fmt.Errorf("ошибка получения обновлений: %v", err)
	}

	// Ищем канал в обновлениях
	for _, update := range updates {
		if update.ChannelPost != nil && update.ChannelPost.Chat != nil {
			if update.ChannelPost.Chat.UserName == strings.TrimPrefix(username, "@") {
				return update.ChannelPost.Chat.ID, nil
			}
		}
		if update.Message != nil && update.Message.Chat != nil {
			if update.Message.Chat.UserName == strings.TrimPrefix(username, "@") {
				return update.Message.Chat.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("канал %s не найден в обновлениях. Убедитесь, что бот добавлен в канал", username)
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
