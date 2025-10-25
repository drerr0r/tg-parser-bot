-- Создание таблицы правил парсинга (если не существует)
CREATE TABLE IF NOT EXISTS parsing_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    source_channel VARCHAR(255) NOT NULL,
    keywords TEXT[], -- массив ключевых слов
    exclude_words TEXT[], -- массив слов-исключений
    media_types VARCHAR(50)[], -- массив типов медиа
    min_text_length INTEGER DEFAULT 0,
    max_text_length INTEGER DEFAULT 0,
    text_replacements JSONB, -- замены текста { "старое": "новое" }
    add_prefix VARCHAR(255) DEFAULT '',
    add_suffix VARCHAR(255) DEFAULT '',
    target_platforms VARCHAR(50)[], -- массив платформ
    check_interval INTEGER NOT NULL DEFAULT 2, -- интервал проверки в минуты
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы постов (если не существует)
CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT REFERENCES parsing_rules(id) ON DELETE CASCADE,
    message_id BIGINT NOT NULL,
    source_channel VARCHAR(255) NOT NULL,
    content TEXT,
    media_type VARCHAR(50),
    media_url TEXT,
    posted_at TIMESTAMP WITH TIME ZONE,
    parsed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_published BOOLEAN DEFAULT FALSE,
    publish_error TEXT,

    -- Уникальность сообщения в канале
    UNIQUE(source_channel, message_id)
);

-- Индексы для оптимизации (если не существуют)
DO $$
BEGIN
    -- Проверяем и создаем индексы только если их нет
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_parsing_rules_source_channel') THEN
        CREATE INDEX idx_parsing_rules_source_channel ON parsing_rules(source_channel);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_parsing_rules_is_active') THEN
        CREATE INDEX idx_parsing_rules_is_active ON parsing_rules(is_active);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_posts_source_channel') THEN
        CREATE INDEX idx_posts_source_channel ON posts(source_channel);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_posts_parsed_at') THEN
        CREATE INDEX idx_posts_parsed_at ON posts(parsed_at);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_posts_is_published') THEN
        CREATE INDEX idx_posts_is_published ON posts(is_published);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_posts_rule_id') THEN
        CREATE INDEX idx_posts_rule_id ON posts(rule_id);
    END IF;
END $$;

-- Функция для обновления updated_at (если не существует)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для обновления updated_at (если не существует)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_parsing_rules_updated_at') THEN
        CREATE TRIGGER update_parsing_rules_updated_at
            BEFORE UPDATE ON parsing_rules
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$;

