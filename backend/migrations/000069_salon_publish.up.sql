CREATE TABLE salon_publish (
    business_id UUID PRIMARY KEY REFERENCES businesses(id) ON DELETE CASCADE,
    content_hash TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
