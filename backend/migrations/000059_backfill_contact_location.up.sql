-- Copy the free-text address and phone that salons already entered in their
-- landing page "contact" section into the new structured businesses columns.
-- This is a one-time transform so existing salons keep their location without
-- user intervention. The city/geo columns are filled later by geocoding.
DO $$
DECLARE
    r RECORD;
    v RECORD;
    new_addr TEXT;
    new_phone TEXT;
BEGIN
    FOR r IN
        SELECT id FROM businesses
        WHERE address_line IS NULL OR phone IS NULL
    LOOP
        new_addr := NULL;
        new_phone := NULL;

        FOR v IN
            SELECT kv.value->>'address' AS address,
                   kv.value->>'phone' AS phone
            FROM sections s
            JOIN pages p ON p.id = s.page_id
            CROSS JOIN LATERAL jsonb_each(s.content) AS kv(key, value)
            WHERE p.business_id = r.id
              AND p.name = 'landing'
              AND s.type = 'contact'
            ORDER BY s.position, kv.key
        LOOP
            IF new_addr IS NULL AND btrim(COALESCE(v.address, '')) <> '' THEN
                new_addr := btrim(v.address);
            END IF;
            IF new_phone IS NULL AND btrim(COALESCE(v.phone, '')) <> '' THEN
                new_phone := btrim(v.phone);
            END IF;
            EXIT WHEN new_addr IS NOT NULL AND new_phone IS NOT NULL;
        END LOOP;

        IF new_addr IS NOT NULL OR new_phone IS NOT NULL THEN
            UPDATE businesses
            SET address_line = COALESCE(address_line, new_addr),
                phone = COALESCE(phone, new_phone),
                updated_at = now()
            WHERE id = r.id;
        END IF;
    END LOOP;
END $$;
