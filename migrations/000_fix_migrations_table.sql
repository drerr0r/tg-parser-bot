-- ИСПРАВЛЕНИЕ ТАБЛИЦЫ MIGRATIONS - УМНАЯ ЗАМЕНА С СОХРАНЕНИЕМ ДАННЫХ
DO $$ 
BEGIN
    -- Проверяем текущую структуру таблицы
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'schema_migrations' 
        AND column_name = 'version' 
        AND data_type = 'bigint'
    ) THEN
        RAISE NOTICE 'Обнаружена неправильная структура schema_migrations (version BIGINT), исправляем...';
        
        -- Создаем временную таблицу с правильной структурой
        CREATE TABLE IF NOT EXISTS schema_migrations_correct (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
        
        -- Переносим данные из старой таблицы если они есть
        IF EXISTS (SELECT 1 FROM schema_migrations) THEN
            INSERT INTO schema_migrations_correct (version, applied_at)
            SELECT version::text, CURRENT_TIMESTAMP 
            FROM schema_migrations
            ON CONFLICT (version) DO NOTHING;
            
            RAISE NOTICE 'Перенесено % записей миграций', FOUND;
        END IF;
        
        -- Заменяем таблицы
        DROP TABLE schema_migrations;
        ALTER TABLE schema_migrations_correct RENAME TO schema_migrations;
        
        RAISE NOTICE 'Таблица schema_migrations исправлена (BIGINT -> VARCHAR)';
    ELSE
        RAISE NOTICE 'Таблица schema_migrations уже имеет правильную структуру';
        
        -- Создаем таблицу если ее нет
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;