-- Structured salon location for local SEO and JSON-LD. All columns are
-- nullable so existing salons are unaffected; owners fill them via the new
-- location endpoint. City is the key local signal for queries like
-- "fade haircut sremska mitrovica".
ALTER TABLE businesses
    ADD COLUMN address_line TEXT,
    ADD COLUMN city TEXT,
    ADD COLUMN postal_code TEXT,
    ADD COLUMN country TEXT,
    ADD COLUMN latitude DOUBLE PRECISION,
    ADD COLUMN longitude DOUBLE PRECISION,
    ADD COLUMN phone TEXT;
