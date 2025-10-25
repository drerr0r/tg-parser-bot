package parser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"

	"github.com/drerr0r/tgparserbot/internal/models"
)

// MTProtoClient клиент для работы с MTProto API
type MTProtoClient struct {
	client  *telegram.Client
	apiID   int
	apiHash string
	phone   string
	session string
	logger  *zap.SugaredLogger
	isAuth  bool
	running bool
}

// NewMTProtoClient создает новый MTProto клиент
func NewMTProtoClient(apiID int, apiHash, phone, session string, logger *zap.SugaredLogger) *MTProtoClient {
	return &MTProtoClient{
		apiID:   apiID,
		apiHash: apiHash,
		phone:   phone,
		session: session,
		logger:  logger,
		isAuth:  false,
		running: false,
	}
}

// Start запускает клиент и выполняет аутентификацию
func (m *MTProtoClient) Start(ctx context.Context) error {
	if m.running {
		return fmt.Errorf("клиент уже запущен")
	}

	m.logger.Info("🔗 Запуск MTProto клиента...")

	client := telegram.NewClient(m.apiID, m.apiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: m.session + ".session",
		},
		Logger: m.logger.Desugar(),
	})

	m.client = client
	m.running = true

	// Запускаем клиент в отдельной горутине
	go func() {
		if err := client.Run(ctx, func(ctx context.Context) error {
			m.logger.Info("✅ Соединение с Telegram установлено")

			// Проверяем статус аутентификации
			authStatus, err := m.client.Auth().Status(ctx)
			if err != nil {
				m.logger.Errorf("❌ Ошибка проверки статуса аутентификации: %v", err)
				return err
			}

			if authStatus.Authorized {
				m.logger.Info("✅ Уже авторизованы в Telegram")
				m.isAuth = true
			} else {
				m.logger.Info("🔐 Аутентификация в Telegram...")
				m.logger.Info("📱 Запрос кода аутентификации...")

				// Если не авторизованы, запрашиваем код
				flow := auth.NewFlow(
					auth.Constant(m.phone, "", auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
						m.logger.Info("📱 Введите код из Telegram:")
						var code string
						fmt.Scanln(&code)
						return code, nil
					})),
					auth.SendCodeOptions{},
				)

				if err := m.client.Auth().IfNecessary(ctx, flow); err != nil {
					m.logger.Errorf("❌ Ошибка аутентификации: %v", err)
					return err
				}

				m.isAuth = true
				m.logger.Info("✅ Успешная аутентификация в Telegram")
			}

			// Держим соединение открытым
			m.logger.Info("🔄 Клиент готов к работе, ожидание запросов...")
			<-ctx.Done()
			return nil
		}); err != nil {
			m.logger.Errorf("❌ Ошибка работы клиента: %v", err)
		}

		m.running = false
		m.logger.Info("🛑 MTProto клиент остановлен")
	}()

	// Ждем инициализации
	time.Sleep(2 * time.Second)
	return nil
}

// Stop останавливает клиент
func (m *MTProtoClient) Stop() {
	m.running = false
}

// GetChannelMessages получает сообщения из канала
func (m *MTProtoClient) GetChannelMessages(ctx context.Context, channel string, limit int) ([]*ParsedMessage, error) {
	if !m.isAuth || m.client == nil {
		return nil, fmt.Errorf("клиент не аутентифицирован или не инициализирован")
	}

	normalizedChannel := strings.TrimPrefix(channel, "@")
	m.logger.Infof("📥 Получение сообщений из канала: %s (лимит: %d)", normalizedChannel, limit)

	api := m.client.API()

	// Ищем канал
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: normalizedChannel,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска канала %s: %v", normalizedChannel, err)
	}

	var channelPeer *tg.Channel
	for _, chat := range resolved.Chats {
		if c, ok := chat.(*tg.Channel); ok {
			channelPeer = c
			break
		}
	}

	if channelPeer == nil {
		return nil, fmt.Errorf("канал %s не найден", normalizedChannel)
	}

	m.logger.Infof("✅ Найден канал: %s (ID: %d)", normalizedChannel, channelPeer.ID)

	// Получаем историю сообщений
	history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelPeer.ID,
			AccessHash: channelPeer.AccessHash,
		},
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории сообщений: %v", err)
	}

	var parsedMessages []*ParsedMessage

	// Обрабатываем сообщения в зависимости от типа результата
	switch result := history.(type) {
	case *tg.MessagesChannelMessages:
		m.logger.Infof("📊 Получено %d сообщений из канала", len(result.Messages))
		for _, msg := range result.Messages {
			parsedMsg, err := m.parseMessage(msg, channel)
			if err != nil {
				m.logger.Warnf("⚠️ Ошибка парсинга сообщения: %v", err)
				continue
			}
			if parsedMsg != nil {
				parsedMessages = append(parsedMessages, parsedMsg)
			}
		}
	case *tg.MessagesMessages:
		m.logger.Infof("📊 Получено %d сообщений", len(result.Messages))
		for _, msg := range result.Messages {
			parsedMsg, err := m.parseMessage(msg, channel)
			if err != nil {
				m.logger.Warnf("⚠️ Ошибка парсинга сообщения: %v", err)
				continue
			}
			if parsedMsg != nil {
				parsedMessages = append(parsedMessages, parsedMsg)
			}
		}
	default:
		return nil, fmt.Errorf("неожиданный тип результата: %T", history)
	}

	m.logger.Infof("✅ Успешно обработано %d сообщений из канала %s", len(parsedMessages), normalizedChannel)
	return parsedMessages, nil
}

// parseMessage парсит сообщение Telegram в нашу структуру
func (m *MTProtoClient) parseMessage(msg tg.MessageClass, channel string) (*ParsedMessage, error) {
	var message *tg.Message

	switch msg := msg.(type) {
	case *tg.Message:
		message = msg
	case *tg.MessageEmpty:
		return nil, nil
	case *tg.MessageService:
		return nil, nil
	default:
		return nil, fmt.Errorf("неподдерживаемый тип сообщения: %T", msg)
	}

	content := message.Message
	if content == "" {
		return nil, nil
	}

	mediaType := models.MediaText
	var mediaURL string

	if message.Media != nil {
		switch media := message.Media.(type) {
		case *tg.MessageMediaPhoto:
			mediaType = models.MediaPhoto
			m.logger.Debugf("📷 Найдено фото в сообщении %d", message.ID)

		case *tg.MessageMediaDocument:
			mediaType = models.MediaDocument
			m.logger.Debugf("📄 Найден документ в сообщении %d", message.ID)

		case *tg.MessageMediaWebPage:
			if media.Webpage != nil {
				if webpage, ok := media.Webpage.(*tg.WebPage); ok {
					mediaURL = webpage.URL
					m.logger.Debugf("🌐 Найдена веб-страница: %s", webpage.URL)
				}
			}
		}
	}

	parsedMsg := &ParsedMessage{
		ID:            int64(message.ID),
		SourceChannel: channel,
		Content:       content,
		MediaType:     mediaType,
		MediaURL:      mediaURL,
		Date:          time.Unix(int64(message.Date), 0),
	}

	m.logger.Debugf("📝 Обработано сообщение %d: %s", message.ID, truncateText(content, 100))
	return parsedMsg, nil
}

// GetNewMessages получает только новые сообщения (после указанного ID)
func (m *MTProtoClient) GetNewMessages(ctx context.Context, channel string, lastMessageID int64) ([]*ParsedMessage, error) {
	if !m.isAuth || m.client == nil {
		return nil, fmt.Errorf("клиент не аутентифицирован или не инициализирован")
	}

	// Получаем последние сообщения и фильтруем те, что новее lastMessageID
	allMessages, err := m.GetChannelMessages(ctx, channel, 50)
	if err != nil {
		return nil, err
	}

	var newMessages []*ParsedMessage
	for _, msg := range allMessages {
		if msg.ID > lastMessageID {
			newMessages = append(newMessages, msg)
			m.logger.Debugf("🆕 Найдено новое сообщение ID %d", msg.ID)
		}
	}

	m.logger.Infof("✅ Найдено %d новых сообщений в канале %s", len(newMessages), channel)
	return newMessages, nil
}

// TestConnection проверяет подключение к Telegram
func (m *MTProtoClient) TestConnection(ctx context.Context) error {
	if !m.isAuth || m.client == nil {
		return fmt.Errorf("клиент не аутентифицирован или не инициализирован")
	}

	_, err := m.client.API().HelpGetConfig(ctx)
	if err != nil {
		return fmt.Errorf("ошибка проверки подключения: %v", err)
	}

	m.logger.Info("✅ Подключение к Telegram работает")
	return nil
}

// Вспомогательная функция для обрезки текста
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
