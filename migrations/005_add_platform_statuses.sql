-- Добавляем отдельные колонки для статусов публикации
ALTER TABLE posts 
ADD COLUMN IF NOT EXISTS published_telegram BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS published_vk BOOLEAN DEFAULT FALSE;

-- Обновляем существующие данные
UPDATE posts 
SET published_telegram = is_published,
    published_vk = is_published
WHERE published_telegram IS NULL OR published_vk IS NULL;

