DROP TABLE IF EXISTS image_links;

ALTER TABLE images DROP COLUMN IF EXISTS business_id;

ALTER TABLE images DROP CONSTRAINT images_storage_payload_check;
ALTER TABLE images ADD CONSTRAINT images_storage_payload_check CHECK (
    (storage = 'minio'    AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 'postgres' AND data       IS NOT NULL AND object_key IS NULL AND url IS NULL) OR
    (storage = 'external' AND url        IS NOT NULL AND object_key IS NULL AND data IS NULL)
) NOT VALID;
