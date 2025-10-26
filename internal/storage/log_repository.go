package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/drerr0r/tgparserbot/internal/models"
)

// LogRepository репозиторий для работы с логами
type LogRepository struct {
	logFilePath string
}

// NewLogRepository создает новый репозиторий логов
func NewLogRepository(logFilePath string) *LogRepository {
	return &LogRepository{
		logFilePath: logFilePath,
	}
}

// LogRecord структура для парсинга JSON логов
type LogRecord struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Logger  string `json:"logger"`
	Caller  string `json:"caller"`
	Message string `json:"msg"`
}

// GetLogs возвращает логи из файла с фильтрацией
func (r *LogRepository) GetLogs(ctx context.Context, filter models.LogFilter) ([]*models.LogEntry, int, error) {
	// Проверяем существует ли файл логов
	if _, err := os.Stat(r.logFilePath); os.IsNotExist(err) {
		// Если файла нет, возвращаем пустой список
		return []*models.LogEntry{}, 0, nil
	}

	file, err := os.Open(r.logFilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка открытия файла логов: %v", err)
	}
	defer file.Close()

	var allLogs []*models.LogEntry
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Парсим JSON строку
		var record LogRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			// Пропускаем строки которые не являются JSON
			continue
		}

		// Парсим время
		timestamp, err := time.Parse(time.RFC3339, record.Time)
		if err != nil {
			// Если не удалось распарсить время, используем текущее
			timestamp = time.Now()
		}

		// Определяем сервис из caller
		service := "system"
		if record.Caller != "" {
			// Извлекаем имя сервиса из пути caller (например: "web/main.go:33" -> "web")
			parts := strings.Split(record.Caller, "/")
			if len(parts) > 0 {
				service = parts[0]
			}
		}

		logEntry := &models.LogEntry{
			ID:        int64(lineNumber),
			Timestamp: timestamp,
			Level:     record.Level,
			Service:   service,
			Message:   record.Message,
			Caller:    record.Caller,
		}

		allLogs = append(allLogs, logEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка чтения файла логов: %v", err)
	}

	// Фильтрация логов
	filteredLogs := r.filterLogs(allLogs, filter)
	total := len(filteredLogs)

	// Сортировка по времени (новые сверху)
	sort.Slice(filteredLogs, func(i, j int) bool {
		return filteredLogs[i].Timestamp.After(filteredLogs[j].Timestamp)
	})

	// Пагинация
	start := filter.Offset
	if start > len(filteredLogs) {
		start = len(filteredLogs)
	}
	end := start + filter.Limit
	if end > len(filteredLogs) {
		end = len(filteredLogs)
	}

	if start >= len(filteredLogs) {
		return []*models.LogEntry{}, total, nil
	}

	return filteredLogs[start:end], total, nil
}

// filterLogs применяет фильтры к логам
func (r *LogRepository) filterLogs(logs []*models.LogEntry, filter models.LogFilter) []*models.LogEntry {
	if filter.Level == "" && filter.Service == "" && filter.Search == "" {
		return logs
	}

	var filtered []*models.LogEntry
	for _, log := range logs {
		// Фильтр по уровню
		if filter.Level != "" && !strings.EqualFold(log.Level, filter.Level) {
			continue
		}

		// Фильтр по сервису
		if filter.Service != "" && !strings.Contains(strings.ToLower(log.Service), strings.ToLower(filter.Service)) {
			continue
		}

		// Поиск по сообщению
		if filter.Search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(filter.Search)) {
			continue
		}

		filtered = append(filtered, log)
	}

	return filtered
}
