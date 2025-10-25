package logger

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger глобальный логгер
var Logger *zap.Logger

// Init инициализирует логгер
func Init(level, format, filePath string) error {
	// Уровень логирования
	logLevel := parseLevel(level)

	// Конфигурация энкодера
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Настройка формата
	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Writers для логов
	var writers []io.Writer

	// Всегда пишем в stdout
	writers = append(writers, os.Stdout)

	// Если указан файл, пишем и в файл
	if filePath != "" {
		// Создаем директорию если нужно
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, file)
	}

	// Создаем core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(io.MultiWriter(writers...))),
		logLevel,
	)

	// Создаем логгер
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// parseLevel парсит уровень логирования
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync синхронизирует логгер
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Sugar возвращает sugared logger
func Sugar() *zap.SugaredLogger {
	if Logger == nil {
		// Fallback на стандартный логгер если основной не инициализирован
		logger, _ := zap.NewProduction()
		return logger.Sugar()
	}
	return Logger.Sugar()
}
