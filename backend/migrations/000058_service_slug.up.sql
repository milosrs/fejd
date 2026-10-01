-- Add a stable, per-business slug to services so public URLs are human-readable
-- (no UUIDs). Backfill existing services by slugifying their name, then make the
-- column required and unique within each business. Keep the transliteration set
-- in sync with slugTransliterations in backend/internal/handler/me_handler.go.

ALTER TABLE services ADD COLUMN slug TEXT;

CREATE OR REPLACE FUNCTION fejd_service_slugify(name text) RETURNS text AS $$
DECLARE
    s text;
BEGIN
    s := lower(btrim(name));
    s := replace(s, 'à','a'); s := replace(s, 'á','a'); s := replace(s, 'â','a');
    s := replace(s, 'ã','a'); s := replace(s, 'ä','a'); s := replace(s, 'å','a');
    s := replace(s, 'ā','a'); s := replace(s, 'ă','a'); s := replace(s, 'ą','a');
    s := replace(s, 'æ','ae');
    s := replace(s, 'ç','c'); s := replace(s, 'ć','c'); s := replace(s, 'č','c');
    s := replace(s, 'ď','d'); s := replace(s, 'đ','d');
    s := replace(s, 'è','e'); s := replace(s, 'é','e'); s := replace(s, 'ê','e');
    s := replace(s, 'ë','e'); s := replace(s, 'ē','e'); s := replace(s, 'ĕ','e');
    s := replace(s, 'ė','e'); s := replace(s, 'ę','e'); s := replace(s, 'ě','e');
    s := replace(s, 'ğ','g');
    s := replace(s, 'ì','i'); s := replace(s, 'í','i'); s := replace(s, 'î','i');
    s := replace(s, 'ï','i'); s := replace(s, 'ĩ','i'); s := replace(s, 'ī','i');
    s := replace(s, 'ĭ','i'); s := replace(s, 'į','i'); s := replace(s, 'ı','i');
    s := replace(s, 'ĺ','l'); s := replace(s, 'ľ','l'); s := replace(s, 'ł','l');
    s := replace(s, 'ñ','n'); s := replace(s, 'ń','n'); s := replace(s, 'ň','n');
    s := replace(s, 'ņ','n');
    s := replace(s, 'ò','o'); s := replace(s, 'ó','o'); s := replace(s, 'ô','o');
    s := replace(s, 'õ','o'); s := replace(s, 'ö','o'); s := replace(s, 'ø','o');
    s := replace(s, 'ō','o'); s := replace(s, 'ő','o');
    s := replace(s, 'œ','oe');
    s := replace(s, 'ŕ','r'); s := replace(s, 'ř','r');
    s := replace(s, 'ś','s'); s := replace(s, 'š','s'); s := replace(s, 'ş','s');
    s := replace(s, 'ș','s');
    s := replace(s, 'ť','t'); s := replace(s, 'ţ','t'); s := replace(s, 'ț','t');
    s := replace(s, 'ù','u'); s := replace(s, 'ú','u'); s := replace(s, 'û','u');
    s := replace(s, 'ü','u'); s := replace(s, 'ũ','u'); s := replace(s, 'ū','u');
    s := replace(s, 'ŭ','u'); s := replace(s, 'ů','u'); s := replace(s, 'ű','u');
    s := replace(s, 'ų','u');
    s := replace(s, 'ý','y'); s := replace(s, 'ÿ','y');
    s := replace(s, 'ź','z'); s := replace(s, 'ż','z'); s := replace(s, 'ž','z');
    s := replace(s, 'ß','ss');
    s := regexp_replace(s, '[^a-z0-9]+', '-', 'g');
    s := btrim(s, '-');
    IF s = '' THEN s := 'service'; END IF;
    IF length(s) > 100 THEN s := left(s, 100); END IF;
    RETURN s;
END $$ LANGUAGE plpgsql;

DO $$
DECLARE
    r RECORD;
    base TEXT;
    candidate TEXT;
    i INT;
    taken BOOLEAN;
BEGIN
    FOR r IN SELECT id, business_id, name FROM services LOOP
        base := fejd_service_slugify(r.name);
        candidate := base;
        i := 2;
        LOOP
            SELECT EXISTS (SELECT 1 FROM services WHERE business_id = r.business_id AND slug = candidate) INTO taken;
            EXIT WHEN NOT taken;
            candidate := base || '-' || i;
            i := i + 1;
        END LOOP;
        UPDATE services SET slug = candidate WHERE id = r.id;
    END LOOP;
END $$;

DROP FUNCTION fejd_service_slugify(text);

ALTER TABLE services ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX uq_services_slug_per_business ON services (business_id, slug);
