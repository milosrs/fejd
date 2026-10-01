ALTER TABLE businesses
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS country,
    DROP COLUMN IF EXISTS postal_code,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS address_line;
