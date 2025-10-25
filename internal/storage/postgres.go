package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/drerr0r/tgparserbot/internal/models"
	"github.com/drerr0r/tgparserbot/pkg/logger"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB представляет подключение к базе данных
type DB struct {
	Pool *pgxpool.Pool
}

// New создает новое подключение к БД
func New(cfg models.DatabaseConfig) (*DB, error) {

	logger.Sugar().Infof("Подключение к БД: host=%s, port=%d, db=%s", cfg.Host, cfg.Port, cfg.Name)

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига БД: %v", err)
	}

	// Настройки пула соединений
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.HealthCheckPeriod = 1 * time.Minute
	poolConfig.MaxConnLifetime = 1 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ошибка ping БД: %v", err)
	}

	return &DB{Pool: pool}, nil
}

// Close закрывает подключение к БД
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// RunMigrations запускает миграции БД
func (db *DB) RunMigrations(migrationsPath string) error {
	logger.Sugar().Infof("🔍 Начинаем применение миграций из: %s", migrationsPath)

	// Автоматически читаем список миграций из папки
	migrations, err := db.readMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файлов миграций: %v", err)
	}

	if len(migrations) == 0 {
		logger.Sugar().Warn("⚠️  В папке миграций не найдено файлов")
		return nil
	}

	logger.Sugar().Infof("📁 Найдено миграций: %d", len(migrations))
	logger.Sugar().Debugf("📋 Список миграций: %v", migrations)

	// Применяем миграции по порядку
	for _, migrationFile := range migrations {
		version := strings.TrimSuffix(migrationFile, ".sql")

		logger.Sugar().Debugf("🔎 Проверяем миграцию: %s", version)

		// Проверяем, применена ли уже эта миграция
		applied, err := db.isMigrationApplied(version)
		if err != nil {
			return fmt.Errorf("ошибка проверки миграции %s: %v", migrationFile, err)
		}

		if applied {
			logger.Sugar().Infof("⏭️  Миграция %s уже применена, пропускаем", migrationFile)
			continue
		}

		logger.Sugar().Infof("🔄 Миграция %s не применена, начинаем применение", migrationFile)

		// Читаем и применяем миграцию
		if err := db.applyMigration(migrationsPath, migrationFile, version); err != nil {
			return err
		}
	}

	logger.Sugar().Info("✅ Все миграции успешно применены")
	return nil
}

// readMigrationFiles читает и сортирует файлы миграций из папки
func (db *DB) readMigrationFiles(migrationsPath string) ([]string, error) {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения папки миграций: %v", err)
	}

	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrations = append(migrations, entry.Name())
		}
	}

	// Сортируем миграции по имени (они должны начинаться с номера)
	sort.Strings(migrations)

	logger.Sugar().Debugf("📋 Найденные миграции: %v", migrations)
	return migrations, nil
}

// isMigrationApplied проверяет применена ли миграция
func (db *DB) isMigrationApplied(version string) (bool, error) {
	// Сначала проверяем существует ли таблица миграций
	var tableExists bool
	err := db.Pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations')",
	).Scan(&tableExists)

	if err != nil {
		return false, fmt.Errorf("ошибка проверки таблицы миграций: %v", err)
	}

	if !tableExists {
		return false, nil
	}

	// Проверяем структуру таблицы - если version имеет тип bigint, нужно исправить
	var columnType string
	err = db.Pool.QueryRow(context.Background(),
		`SELECT data_type FROM information_schema.columns 
         WHERE table_name = 'schema_migrations' AND column_name = 'version'`,
	).Scan(&columnType)

	if err != nil {
		return false, fmt.Errorf("ошибка проверки структуры таблицы: %v", err)
	}

	// Если тип BIGINT, таблица неправильная - считаем что миграция не применена
	if columnType == "bigint" {
		logger.Sugar().Warn("⚠️  Обнаружена неправильная структура schema_migrations (version BIGINT)")
		return false, nil
	}

	// Проверяем применена ли конкретная миграция
	var applied bool
	err = db.Pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
		version,
	).Scan(&applied)

	if err != nil {
		return false, fmt.Errorf("ошибка проверки миграции %s: %v", version, err)
	}

	return applied, nil
}

// applyMigration применяет одну миграцию
func (db *DB) applyMigration(migrationsPath, migrationFile, version string) error {
	// Читаем файл миграции
	filePath := filepath.Join(migrationsPath, migrationFile)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения миграции %s: %v", migrationFile, err)
	}

	logger.Sugar().Infof("📝 Применяем миграцию: %s", migrationFile)

	// Выполняем миграцию в транзакции
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции для %s: %v", migrationFile, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	// Выполняем SQL миграции
	if _, err := tx.Exec(context.Background(), string(content)); err != nil {
		return fmt.Errorf("ошибка выполнения миграции %s: %v", migrationFile, err)
	}

	// Отмечаем миграцию как примененную
	markQuery := "INSERT INTO schema_migrations (version) VALUES ($1)"
	if _, err := tx.Exec(context.Background(), markQuery, version); err != nil {
		return fmt.Errorf("ошибка отметки миграции %s: %v", migrationFile, err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("ошибка коммита транзакции для %s: %v", migrationFile, err)
	}

	logger.Sugar().Infof("✅ Применена миграция: %s", migrationFile)
	return nil
}

// HealthCheck проверяет доступность БД
func (db *DB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Pool.Ping(ctx)
}

// CheckColumnsExist проверяет существование необходимых колонок в таблице posts
func (db *DB) CheckColumnsExist() bool {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = 'posts'
		AND column_name IN (
			'id', 'rule_id', 'message_id', 'source_channel', 'content', 
			'media_type', 'media_url', 'posted_at', 'parsed_at',
			'published_telegram', 'published_vk', 'publish_error'
		)
	`

	var count int
	err := db.Pool.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return false
	}

	// Должно быть 12 колонок (все необходимые)
	return count == 12
}

// CheckMigrationsApplied проверяет применены ли все миграции
func (db *DB) CheckMigrationsApplied() bool {
	// Проверяем существует ли таблица миграций
	var tableExists bool
	err := db.Pool.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations')",
	).Scan(&tableExists)

	if err != nil || !tableExists {
		return false
	}

	// Получаем список всех файлов миграций
	migrations, err := db.readMigrationFiles("./migrations")
	if err != nil || len(migrations) == 0 {
		return false
	}

	// Проверяем что все миграции применены
	for _, migrationFile := range migrations {
		version := strings.TrimSuffix(migrationFile, ".sql")

		var applied bool
		err := db.Pool.QueryRow(context.Background(),
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)",
			version,
		).Scan(&applied)

		if err != nil || !applied {
			return false
		}
	}

	return true
}
