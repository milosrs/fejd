-- Rename any existing salon whose slug collides with a reserved subdomain, so
-- infra hosts (www, api, auth, ...) are never shadowed by a salon. Keep in sync
-- with the reservedSubdomains set in backend/internal/handler/me_handler.go.
DO $$
DECLARE
    r RECORD;
    candidate TEXT;
    i INT;
    taken BOOLEAN;
BEGIN
    FOR r IN
        SELECT id, slug FROM businesses
        WHERE slug IN (
            'www', 'api', 'app', 'auth', 'keycloak',
            'admin', 'm', 'static', 'cdn', 'smtp',
            'mail', 'help', 'support', 'status', 'docs',
            'blog', 'staging', 'dev', 'test'
        )
    LOOP
        i := 2;
        LOOP
            candidate := r.slug || '-' || i;
            SELECT EXISTS (SELECT 1 FROM businesses WHERE slug = candidate) INTO taken;
            EXIT WHEN NOT taken;
            i := i + 1;
        END LOOP;
        UPDATE businesses SET slug = candidate WHERE id = r.id;
    END LOOP;
END $$;
