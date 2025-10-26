package models

import (
	"time"
)

// LogEntry - запись лога
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	Caller    string    `json:"caller,omitempty"`
}

// LogFilter - фильтр для логов
type LogFilter struct {
	Level   string `json:"level"`
	Service string `json:"service"`
	Search  string `json:"search"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
}
