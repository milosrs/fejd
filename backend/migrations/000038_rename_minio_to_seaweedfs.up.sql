-- Rename the image storage backend identifier minio -> seaweedfs.
-- SeaweedFS speaks the S3 protocol, so nothing about the payload shape
-- changes; only the storage enum label and its payload CHECK are renamed.

ALTER TABLE images DROP CONSTRAINT images_storage_payload_check;

UPDATE images SET storage = 'seaweedfs' WHERE storage = 'minio';

ALTER TABLE images ADD CONSTRAINT images_storage_payload_check CHECK (
    (storage = 'seaweedfs' AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 's3'       AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 'postgres' AND data       IS NOT NULL AND object_key IS NULL AND url IS NULL) OR
    (storage = 'external' AND url        IS NOT NULL AND object_key IS NULL AND data IS NULL)
);
