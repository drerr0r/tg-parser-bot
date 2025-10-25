-- Добавляем колонку check_interval в таблицу parsing_rules
ALTER TABLE parsing_rules ADD COLUMN IF NOT EXISTS check_interval INTEGER NOT NULL DEFAULT 2;

-- Обновляем существующие записи, устанавливая значение по умолчанию
UPDATE parsing_rules SET check_interval = 2 WHERE check_interval IS NULL;

-- Обновляем тестовые данные с указанием интервала проверки
UPDATE parsing_rules SET 
    check_interval = 2,
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'NewsWorldTrading - Финансовые новости';

