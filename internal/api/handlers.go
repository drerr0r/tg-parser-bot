package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/internal/storage"
	"github.com/drerr0r/tgparserbot/internal/utils"
	"go.uber.org/zap"
)

type Handlers struct {
	ruleRepo *storage.RuleRepository
	postRepo *storage.PostRepository
	userRepo *storage.UserRepository
	logRepo  *storage.LogRepository
	logger   *zap.SugaredLogger
	cfg      *models.Config
}

func NewHandlers(ruleRepo *storage.RuleRepository, postRepo *storage.PostRepository, userRepo *storage.UserRepository, logRepo *storage.LogRepository, logger *zap.SugaredLogger, cfg *models.Config) *Handlers {
	return &Handlers{
		ruleRepo: ruleRepo,
		postRepo: postRepo,
		userRepo: userRepo,
		logRepo:  logRepo,
		logger:   logger,
		cfg:      cfg,
	}
}

// HealthCheck handler
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "ok",
		"service": "tg-parser-bot",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRules возвращает все правила
func (h *Handlers) GetRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rules, err := h.ruleRepo.List(ctx, 100, 0)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка получения правил: %v", err)
		return
	}

	h.sendJSON(w, http.StatusOK, rules)
}

// CreateRule создает новое правило
func (h *Handlers) CreateRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rule models.ParsingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.sendError(w, http.StatusBadRequest, "Ошибка парсинга JSON: %v", err)
		return
	}

	// Валидация
	if err := rule.Validate(); err != nil {
		h.sendError(w, http.StatusBadRequest, "Ошибка валидации: %v", err)
		return
	}

	// Создаем правило
	if err := h.ruleRepo.Create(ctx, &rule); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка создания правила: %v", err)
		return
	}

	h.logger.Infof("✅ Создано новое правило: %s для канала %s", rule.Name, rule.SourceChannel)
	h.sendJSON(w, http.StatusCreated, rule)
}

// UpdateRule обновляет правило
func (h *Handlers) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем ID из path parameter {id}
	idStr := r.PathValue("id")
	if idStr == "" {
		h.sendError(w, http.StatusBadRequest, "ID правила не указан")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Неверный ID правила: %v", err)
		return
	}

	var rule models.ParsingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.sendError(w, http.StatusBadRequest, "Ошибка парсинга JSON: %v", err)
		return
	}

	// Устанавливаем ID из URL
	rule.ID = id

	// Валидация
	if err := rule.Validate(); err != nil {
		h.sendError(w, http.StatusBadRequest, "Ошибка валидации: %v", err)
		return
	}

	// Обновляем правило
	if err := h.ruleRepo.Update(ctx, &rule); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка обновления правила: %v", err)
		return
	}

	h.logger.Infof("✏️ Обновлено правило ID %d: %s", rule.ID, rule.Name)
	h.sendJSON(w, http.StatusOK, rule)
}

// DeleteRule удаляет правило
func (h *Handlers) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем ID из path parameter {id}
	idStr := r.PathValue("id")
	if idStr == "" {
		h.sendError(w, http.StatusBadRequest, "ID правила не указан")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Неверный ID правила: %v", err)
		return
	}

	if err := h.ruleRepo.Delete(ctx, id); err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка удаления правила: %v", err)
		return
	}

	h.logger.Infof("🗑️ Удалено правило ID %d", id)
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Правило удалено"})
}

// GetPosts возвращает посты
func (h *Handlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем параметры пагинации
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	// Используем новый метод для всех постов
	posts, err := h.postRepo.GetPosts(ctx, limit, offset)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка получения постов: %v", err)
		return
	}

	// Если нет постов, возвращаем пустой массив вместо null
	if posts == nil {
		posts = []*models.Post{}
	}

	h.sendJSON(w, http.StatusOK, posts)
}

// GetStats возвращает статистику
func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем все правила
	allRules, err := h.ruleRepo.List(ctx, 1000, 0)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка получения статистики правил: %v", err)
		return
	}

	// Считаем активные/неактивные правила
	activeRules := 0
	inactiveRules := 0
	for _, rule := range allRules {
		if rule.IsActive {
			activeRules++
		} else {
			inactiveRules++
		}
	}

	// Получаем все посты для статистики
	allPosts, err := h.postRepo.GetPosts(ctx, 10000, 0)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка получения статистики постов: %v", err)
		return
	}

	// Считаем статистику по постам
	totalPosts := len(allPosts)
	telegramPosts := 0
	vkPosts := 0
	successPosts := 0
	failedPosts := 0

	for _, post := range allPosts {
		if post.PublishedTelegram {
			telegramPosts++
		}
		if post.PublishedVK {
			vkPosts++
		}
		if post.PublishError != "" {
			failedPosts++
		} else {
			successPosts++
		}
	}

	stats := map[string]interface{}{
		// Основная статистика
		"rules_count":    len(allRules),
		"posts_count":    totalPosts,
		"telegram_posts": telegramPosts,
		"vk_posts":       vkPosts,

		// Детальная статистика
		"active_rules":   activeRules,
		"inactive_rules": inactiveRules,
		"success_posts":  successPosts,
		"failed_posts":   failedPosts,

		"service": "tg-parser-bot",
		"status":  "running",
	}

	h.sendJSON(w, http.StatusOK, stats)
}

// Login обработчик входа
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из БД
	user, err := h.userRepo.GetByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Errorf("Ошибка получения пользователя: %v", err)
		writeJSONError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if user == nil {
		writeJSONError(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		writeJSONError(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := GenerateJWTToken(user, h.cfg.Auth.JWTSecret, h.cfg.Auth.JWTDuration)
	if err != nil {
		h.logger.Errorf("Ошибка генерации токена: %v", err)
		writeJSONError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// GetCurrentUser возвращает текущего пользователя
func (h *Handlers) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Используем новые функции для получения из контекста
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("Ошибка получения пользователя: %v", err)
		writeJSONError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	if user == nil {
		writeJSONError(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Вспомогательные методы

func (h *Handlers) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) sendError(w http.ResponseWriter, status int, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	h.logger.Error(message)

	errorResponse := map[string]string{
		"error":  message,
		"status": "error",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse)
}

// NotImplemented временный handler
func (h *Handlers) NotImplemented(w http.ResponseWriter, r *http.Request) {
	h.sendError(w, http.StatusNotImplemented, "Функционал в разработке")
}

// ServeFrontend обслуживает фронтенд или возвращает информационное сообщение
func (h *Handlers) ServeFrontend(w http.ResponseWriter, r *http.Request) {
	// Если запрос к API - пропускаем
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.sendError(w, http.StatusNotFound, "API endpoint not found")
		return
	}

	// Проверяем разные возможные пути
	possiblePaths := []string{
		"./web/frontend/dist",
		"web/frontend/dist",
		"/app/web/frontend/dist",
	}

	var actualPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path + "/index.html"); err == nil {
			actualPath = path
			break
		}
	}

	if actualPath == "" {
		// Фронтенд не найден - возвращаем API информацию
		h.sendJSON(w, http.StatusOK, map[string]string{
			"message":  "TG Parser Bot API",
			"status":   "running",
			"frontend": "not built",
			"api_docs": "/docs",
		})
		return
	}

	// Определяем какой файл отдавать
	filePath := actualPath + r.URL.Path
	if r.URL.Path == "/" {
		filePath = actualPath + "/index.html"
	}

	// Проверяем существует ли запрашиваемый файл
	if _, err := os.Stat(filePath); err == nil {
		http.ServeFile(w, r, filePath)
		return
	}

	// Для SPA - все неизвестные пути ведут на index.html
	http.ServeFile(w, r, actualPath+"/index.html")
}

// GetLogs возвращает логи системы
func (h *Handlers) GetLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Парсим параметры запроса
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")
	search := r.URL.Query().Get("search")

	if limit == 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	// Создаем фильтр
	filter := models.LogFilter{
		Level:   level,
		Service: service,
		Search:  search,
		Limit:   limit,
		Offset:  offset,
	}

	// Получаем логи из файла
	logs, total, err := h.logRepo.GetLogs(ctx, filter)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Ошибка чтения логов: %v", err)
		return
	}

	response := map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	h.sendJSON(w, http.StatusOK, response)
}
