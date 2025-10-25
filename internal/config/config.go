package config

import (
	"fmt"
	"os"

	"github.com/drerr0r/tgparserbot/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadConfig загружает конфигурацию из YAML файла
func LoadConfig(path string) (*models.Config, error) {
	config := &models.Config{}

	// Читаем файл
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига %s: %v", path, err)
	}

	// Парсим YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %v", err)
	}

	// Устанавливаем значения по умолчанию
	setDefaults(config)

	return config, nil
}

// setDefaults устанавливает значения по умолчанию
func setDefaults(config *models.Config) {
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Logger.Level == "" {
		config.Logger.Level = "info"
	}
	if config.Logger.Format == "" {
		config.Logger.Format = "json"
	}
	if config.VK.Version == "" {
		config.VK.Version = "5.131"
	}
}

// Validate проверяет обязательные поля конфигурации
func Validate(config *models.Config) error {
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Telegram.APIID == 0 {
		return fmt.Errorf("telegram api_id is required")
	}
	if config.Telegram.APIHash == "" {
		return fmt.Errorf("telegram api_hash is required")
	}
	if config.Telegram.Phone == "" {
		return fmt.Errorf("telegram phone is required")
	}

	return nil
}
