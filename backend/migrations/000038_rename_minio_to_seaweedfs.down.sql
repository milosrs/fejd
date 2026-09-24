ALTER TABLE images DROP CONSTRAINT images_storage_payload_check;

UPDATE images SET storage = 'minio' WHERE storage = 'seaweedfs';

ALTER TABLE images ADD CONSTRAINT images_storage_payload_check CHECK (
    (storage = 'minio'    AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 's3'       AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 'postgres' AND data       IS NOT NULL AND object_key IS NULL AND url IS NULL) OR
    (storage = 'external' AND url        IS NOT NULL AND object_key IS NULL AND data IS NULL)
);
