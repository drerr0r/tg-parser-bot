
-- Добавляем тестовое правило для канала NewsWorldTrading
INSERT INTO parsing_rules (
    name, 
    source_channel, 
    keywords, 
    exclude_words, 
    media_types,
    min_text_length,
    max_text_length, 
    text_replacements,
    add_prefix,
    add_suffix,
    target_platforms,
    check_interval,
    is_active
) VALUES (
    'NewsWorldTrading - Финансовые новости',
    't.me/NewsWorldTrading',
    ARRAY['новости', 'финансы', 'трейдинг', 'рынок', 'акции', 'биржа', 'инвестиции', 'экономика'],
    ARRAY['реклама', 'спам', 'куплю', 'продам', 'курс', 'гарант'],
    ARRAY['text', 'photo'],
    10,
    1000,
    '{"гарант": "возможность", "купить": "рассмотреть"}'::jsonb,
    '📈 ',
    ' #финансы',
    ARRAY['telegram', 'vk'],
    2, -- интервал проверки 2 минуты
    TRUE
) ON CONFLICT DO NOTHING;

