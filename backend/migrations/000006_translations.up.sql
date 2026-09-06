-- App UI labels, stored per locale so the frontend fetches them from the API
-- instead of shipping hardcoded strings. Locales are data, not schema — we
-- start with English (en) and Serbian (rs).

CREATE TABLE translations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key        TEXT NOT NULL,
    locale     TEXT NOT NULL,
    value      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_translations_key_locale ON translations (key, locale);

INSERT INTO translations (key, locale, value) VALUES
    ('landing.empty.title', 'en', 'This page isn''t set up yet'),
    ('landing.empty.body',  'en', 'The salon hasn''t published any content yet. Check back soon.'),
    ('landing.empty.title', 'rs', 'Ova stranica još nije podešena'),
    ('landing.empty.body',  'rs', 'Salon još uvek nije objavio sadržaj. Vratite se uskoro.');
