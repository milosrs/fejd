-- Landing page structure: a business has pages; each page has ordered content
-- sections. For now the only page is the landing page (name 'landing').

CREATE TABLE pages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    position    INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_pages_business_name ON pages (business_id, name);

-- Content is a locale-keyed JSONB object: the top-level keys are locale codes
-- (e.g. "en", "rs") and each value is that section type's localized fields.
-- Adding a language is just another key — no migration. Non-localized concerns
-- (e.g. gallery image urls) live inside the per-locale object for now to keep
-- the contract uniform.
CREATE TABLE sections (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id    UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('hero','about','gallery','contact')),
    content    JSONB NOT NULL DEFAULT '{}'::jsonb,
    position   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sections_page ON sections (page_id, position);
