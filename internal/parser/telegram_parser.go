package parser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/publisher"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"go.uber.org/zap"
)

// TelegramParser парсер Telegram каналов с реальным MTProto
type TelegramParser struct {
	storage        *storage.DB
	ruleRepo       *storage.RuleRepository
	postRepo       *storage.PostRepository
	multiPublisher *publisher.MultiPublisher
	mtprotoClient  *MTProtoClient
	logger         *zap.SugaredLogger
	isRunning      bool
	cancelFunc     context.CancelFunc
	lastMessageIDs map[string]int64 // Храним последние ID сообщений по каналам
}

// NewTelegramParser создает новый парсер
func NewTelegramParser(
	storage *storage.DB,
	ruleRepo *storage.RuleRepository,
	postRepo *storage.PostRepository,
	multiPublisher *publisher.MultiPublisher,
	mtprotoClient *MTProtoClient,
	logger *zap.SugaredLogger,
) *TelegramParser {
	return &TelegramParser{
		storage:        storage,
		ruleRepo:       ruleRepo,
		postRepo:       postRepo,
		multiPublisher: multiPublisher,
		mtprotoClient:  mtprotoClient,
		logger:         logger,
		isRunning:      false,
		lastMessageIDs: make(map[string]int64),
	}
}

// Start запускает парсинг каналов
func (p *TelegramParser) Start(ctx context.Context) error {
	if p.isRunning {
		return fmt.Errorf("парсер уже запущен")
	}

	p.logger.Info("🚀 Запуск Telegram парсера каналов с MTProto...")

	// Запускаем MTProto клиент
	if err := p.mtprotoClient.Start(ctx); err != nil {
		return fmt.Errorf("ошибка запуска MTProto клиента: %v", err)
	}

	// Ждем немного для инициализации
	time.Sleep(3 * time.Second)

	// Загружаем активные правила
	rules, err := p.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return fmt.Errorf("ошибка загрузки правил: %v", err)
	}

	if len(rules) == 0 {
		p.logger.Warn("⚠️ Нет активных правил для парсинга")
		return nil
	}

	p.logger.Infof("📋 Загружено %d активных правил", len(rules))

	// Создаем контекст с отменой
	ctx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel
	p.isRunning = true

	// Инициализируем lastMessageIDs
	for _, rule := range rules {
		lastPost, err := p.postRepo.GetLastMessageID(ctx, rule.SourceChannel)
		if err == nil && lastPost != nil {
			p.lastMessageIDs[rule.SourceChannel] = lastPost.MessageID
		} else {
			p.lastMessageIDs[rule.SourceChannel] = 0
		}
	}

	// Запускаем мониторинг для каждого правила
	for _, rule := range rules {
		p.logger.Infof("🎯 Запуск мониторинга для канала: %s", rule.SourceChannel)
		go p.monitorChannel(ctx, rule)
	}

	p.logger.Info("✅ Telegram парсер успешно запущен")
	return nil
}

// Stop останавливает парсинг
func (p *TelegramParser) Stop() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	// MTProto клиент автоматически закрывается при отмене контекста
	p.isRunning = false
	p.logger.Info("🛑 Парсер остановлен")
}

// monitorChannel мониторит конкретный канал
func (p *TelegramParser) monitorChannel(ctx context.Context, rule *models.ParsingRule) {
	channelDisplay := p.getChannelDisplayName(rule.SourceChannel)
	p.logger.Infof("🔍 Начало мониторинга канала: %s", channelDisplay)

	// Сначала проверяем исторические сообщения
	if err := p.checkHistoricalMessages(ctx, rule); err != nil {
		p.logger.Errorf("❌ Ошибка проверки исторических сообщений: %v", err)
	}

	// Настраиваем интервал проверки
	interval := p.getCheckInterval(rule)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.logger.Infof("⏰ Мониторинг канала %s с интервалом %v", channelDisplay, interval)

	for {
		select {
		case <-ctx.Done():
			p.logger.Infof("🛑 Остановка мониторинга канала: %s", channelDisplay)
			return
		case <-ticker.C:
			if err := p.checkNewMessages(ctx, rule); err != nil {
				p.logger.Errorf("❌ Ошибка проверки сообщений в канале %s: %v", channelDisplay, err)
			}
		}
	}
}

// checkHistoricalMessages проверяет исторические сообщения
func (p *TelegramParser) checkHistoricalMessages(ctx context.Context, rule *models.ParsingRule) error {
	channelDisplay := p.getChannelDisplayName(rule.SourceChannel)
	p.logger.Infof("📚 Проверка исторических сообщений для канала: %s", channelDisplay)

	// Получаем реальные сообщения с канала через MTProto
	messages, err := p.mtprotoClient.GetChannelMessages(ctx, p.normalizeChannel(rule.SourceChannel), 20)
	if err != nil {
		return fmt.Errorf("❌ Ошибка получения сообщений: %v", err)
	}

	p.logger.Infof("📥 Получено %d сообщений с канала %s", len(messages), channelDisplay)

	// Обрабатываем сообщения
	processedCount := 0
	for _, msg := range messages {
		if err := p.processMessage(ctx, rule, msg); err != nil {
			p.logger.Errorf("❌ Ошибка обработки сообщения: %v", err)
			continue
		}
		processedCount++

		// Обновляем последний ID сообщения
		if msg.ID > p.lastMessageIDs[rule.SourceChannel] {
			p.lastMessageIDs[rule.SourceChannel] = msg.ID
		}

		// Небольшая задержка между обработкой
		time.Sleep(100 * time.Millisecond)
	}

	p.logger.Infof("✅ Обработано %d/%d сообщений с канала %s", processedCount, len(messages), channelDisplay)
	return nil
}

// checkNewMessages проверяет новые сообщения
func (p *TelegramParser) checkNewMessages(ctx context.Context, rule *models.ParsingRule) error {
	channelDisplay := p.getChannelDisplayName(rule.SourceChannel)
	p.logger.Debugf("🔄 Проверка новых сообщений в канале: %s", channelDisplay)

	lastMessageID := p.lastMessageIDs[rule.SourceChannel]

	// Получаем только новые сообщения через MTProto
	messages, err := p.mtprotoClient.GetNewMessages(ctx, p.normalizeChannel(rule.SourceChannel), lastMessageID)
	if err != nil {
		return fmt.Errorf("ошибка получения новых сообщений: %v", err)
	}

	if len(messages) > 0 {
		p.logger.Infof("🆕 Найдено %d новых сообщений в канале %s", len(messages), channelDisplay)
	}

	for _, msg := range messages {
		if err := p.processMessage(ctx, rule, msg); err != nil {
			p.logger.Errorf("❌ Ошибка обработки нового сообщения: %v", err)
		} else {
			// Обновляем последний ID сообщения
			if msg.ID > p.lastMessageIDs[rule.SourceChannel] {
				p.lastMessageIDs[rule.SourceChannel] = msg.ID
			}
		}
	}

	return nil
}

// processMessage обрабатывает сообщение (остается без изменений)
func (p *TelegramParser) processMessage(ctx context.Context, rule *models.ParsingRule, msg *ParsedMessage) error {
	// Проверяем, не обрабатывали ли мы уже это сообщение
	existingPost, err := p.postRepo.GetByMessageID(ctx, rule.SourceChannel, msg.ID)
	if err != nil {
		return fmt.Errorf("ошибка проверки существующего поста: %v", err)
	}

	if existingPost != nil {
		p.logger.Debugf("⚠️ Сообщение %d уже обработано", msg.ID)
		return nil
	}

	// Применяем фильтры правила
	if !p.applyFilters(rule, msg) {
		p.logger.Debugf("🚫 Сообщение %d не прошло фильтры", msg.ID)
		return nil
	}

	// Применяем трансформации
	transformedContent := rule.ApplyTransformations(msg.Content)

	// Создаем пост
	post := models.NewPost(rule.ID, msg.ID, rule.SourceChannel, transformedContent, msg.MediaType)
	post.MediaURL = msg.MediaURL
	post.PostedAt = msg.Date
	// По умолчанию оба статуса false (уже установлено в NewPost)
	post.PublishedTelegram = false
	post.PublishedVK = false

	if err := post.Validate(); err != nil {
		return fmt.Errorf("ошибка валидации поста: %v", err)
	}

	// Сохраняем в БД
	if err := p.postRepo.Create(ctx, post); err != nil {
		return fmt.Errorf("ошибка сохранения поста: %v", err)
	}

	p.logger.Infof("💾 Сообщение %d сохранено как пост ID %d", msg.ID, post.ID)

	// Публикуем пост
	p.logger.Infof("📤 Начинаем публикацию поста ID %d", post.ID)
	if err := p.multiPublisher.Publish(ctx, post, rule); err != nil {
		p.logger.Errorf("❌ Ошибка публикации поста %d: %v", post.ID, err)
		return fmt.Errorf("ошибка публикации: %v", err)
	}

	p.logger.Infof("✅ Пост ID %d успешно опубликован", post.ID)
	return nil
}

// applyFilters применяет фильтры правила к сообщению (остается без изменений)
func (p *TelegramParser) applyFilters(rule *models.ParsingRule, msg *ParsedMessage) bool {
	// Проверка ключевых слов
	if !rule.MatchesKeywords(msg.Content) {
		return false
	}

	// Проверка слов-исключений
	if rule.ContainsExcludedWords(msg.Content) {
		return false
	}

	// Проверка типа медиа
	if !rule.SupportsMediaType(msg.MediaType) {
		return false
	}

	// Проверка длины текста
	if !rule.TextLengthValid(msg.Content) {
		return false
	}

	return true
}

// normalizeChannel нормализует формат канала (остается без изменений)
func (p *TelegramParser) normalizeChannel(channel string) string {
	channel = strings.TrimSpace(channel)
	channel = strings.TrimPrefix(channel, "https://")
	channel = strings.TrimPrefix(channel, "http://")
	channel = strings.TrimPrefix(channel, "t.me/")
	channel = strings.TrimPrefix(channel, "@")
	return "@" + channel
}

// getChannelDisplayName возвращает отображаемое имя канала (остается без изменений)
func (p *TelegramParser) getChannelDisplayName(channel string) string {
	normalized := p.normalizeChannel(channel)
	return fmt.Sprintf("%s (%s)", normalized, channel)
}

// getCheckInterval возвращает интервал проверки
func (p *TelegramParser) getCheckInterval(rule *models.ParsingRule) time.Duration {
	// Используем интервал из правила, по умолчанию 2 минуты
	interval := rule.CheckInterval
	if interval <= 0 {
		interval = 2
	}
	return time.Duration(interval) * time.Minute
}

// ParsedMessage структура распаршенного сообщения (остается без изменений)
type ParsedMessage struct {
	ID            int64
	SourceChannel string
	Content       string
	MediaType     models.MediaType
	MediaURL      string
	Date          time.Time
}
