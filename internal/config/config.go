package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/drerr0r/tgparserbot/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadConfig универсальная загрузка конфигурации
// Сначала пробует файл, если нет - только env variables
func LoadConfig(path string) (*models.Config, error) {
	config := &models.Config{}

	// Пробуем загрузить из файла
	data, err := os.ReadFile(path)
	if err != nil {
		// Файла нет - используем только env variables
		fmt.Printf("Config file not found, using environment variables only\n")
	} else {
		// Файл есть - парсим YAML
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("ошибка парсинга YAML: %v", err)
		}
	}

	// Устанавливаем значения по умолчанию
	setDefaults(config)

	// Переопределяем из переменных окружения (ВСЕГДА имеет высший приоритет)
	loadFromEnv(config)

	return config, nil
}

// loadFromEnv переопределяет конфиг из переменных окружения
func loadFromEnv(config *models.Config) {

	fmt.Printf("=== DEBUG loadFromEnv ===\n")
	fmt.Printf("PGHOST='%s'\n", os.Getenv("PGHOST"))
	fmt.Printf("PGDATABASE='%s'\n", os.Getenv("PGDATABASE"))
	fmt.Printf("PGUSER='%s'\n", os.Getenv("PGUSER"))
	fmt.Printf("PGPORT='%s'\n", os.Getenv("PGPORT"))
	fmt.Printf("DB_HOST='%s'\n", os.Getenv("DB_HOST"))
	fmt.Printf("DB_NAME='%s'\n", os.Getenv("DB_NAME"))
	fmt.Printf("=========================\n")

	// Database - поддерживаем оба формата: Railway (PG*) и стандартный (DB_*)
	if host := os.Getenv("PGHOST"); host != "" {
		config.Database.Host = host
	} else if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}

	if port := os.Getenv("PGPORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	} else if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}

	if name := os.Getenv("PGDATABASE"); name != "" {
		config.Database.Name = name
	} else if name := os.Getenv("DB_NAME"); name != "" {
		config.Database.Name = name
	}

	if user := os.Getenv("PGUSER"); user != "" {
		config.Database.User = user
	} else if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}

	if password := os.Getenv("PGPASSWORD"); password != "" {
		config.Database.Password = password
	} else if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}

	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	} else {
		// Для Railway используем require по умолчанию
		config.Database.SSLMode = "require"
	}

	// Telegram
	if apiID := os.Getenv("TG_API_ID"); apiID != "" {
		if id, err := strconv.Atoi(apiID); err == nil {
			config.Telegram.APIID = id
		}
	}
	if apiHash := os.Getenv("TG_API_HASH"); apiHash != "" {
		config.Telegram.APIHash = apiHash
	}
	if phone := os.Getenv("TG_PHONE"); phone != "" {
		config.Telegram.Phone = phone
	}
	if botToken := os.Getenv("TG_BOT_TOKEN"); botToken != "" {
		config.Telegram.BotToken = botToken
	}
	if targetChannel := os.Getenv("TG_TARGET_CHANNEL"); targetChannel != "" {
		config.Telegram.TargetChannel = targetChannel
	}

	// VK
	if accessToken := os.Getenv("VK_ACCESS_TOKEN"); accessToken != "" {
		config.VK.AccessToken = accessToken
	}
	if groupID := os.Getenv("VK_GROUP_ID"); groupID != "" {
		if id, err := strconv.Atoi(groupID); err == nil {
			config.VK.GroupID = id
		}
	}

	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	// Auth
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Auth.JWTSecret = jwtSecret
	}
	if jwtDuration := os.Getenv("JWT_DURATION"); jwtDuration != "" {
		if duration, err := strconv.Atoi(jwtDuration); err == nil {
			config.Auth.JWTDuration = duration
		}
	}

	// Logger
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logger.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logger.Format = format
	}
	if filePath := os.Getenv("LOG_FILE_PATH"); filePath != "" {
		config.Logger.FilePath = filePath
	}

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
	if config.Auth.JWTDuration == 0 {
		config.Auth.JWTDuration = 24
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
