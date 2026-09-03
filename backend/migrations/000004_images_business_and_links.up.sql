-- Image ownership + polymorphic linking.
-- business_id is nullable so the migration never fails on pre-existing images
-- rows; the app always sets it on insert. ON DELETE RESTRICT forces any future
-- business deletion to remove images (including object-store bytes) explicitly
-- rather than silently leaking them.

ALTER TABLE images ADD COLUMN business_id UUID REFERENCES businesses(id) ON DELETE RESTRICT;

-- Allow the s3 backend alongside minio/postgres/external.
ALTER TABLE images DROP CONSTRAINT images_storage_payload_check;
ALTER TABLE images ADD CONSTRAINT images_storage_payload_check CHECK (
    (storage = 'minio'    AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 's3'       AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
    (storage = 'postgres' AND data       IS NOT NULL AND object_key IS NULL AND url IS NULL) OR
    (storage = 'external' AND url        IS NOT NULL AND object_key IS NULL AND data IS NULL)
);

CREATE TABLE image_links (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id    UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL,
    entity_id   UUID NOT NULL,
    purpose     TEXT NOT NULL DEFAULT '',
    visibility  TEXT NOT NULL DEFAULT 'private'
                CHECK (visibility IN ('public','private')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_image_links_entity ON image_links (entity_type, entity_id);
CREATE INDEX idx_image_links_image  ON image_links (image_id);

-- Singleton purposes: one link per (entity, purpose). Re-upload replaces the
-- image_id in place rather than adding a second link.
CREATE UNIQUE INDEX idx_image_links_singleton
    ON image_links (entity_type, entity_id, purpose)
    WHERE purpose IN ('hero','logo','background','avatar','picture');
