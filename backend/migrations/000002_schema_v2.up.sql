-- Overlap prevention needs btree_gist (used by two exclusion constraints).
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Service: description + backend-agnostic image reference.
ALTER TABLE services ADD COLUMN description TEXT;

CREATE TABLE images (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage      TEXT NOT NULL,
    object_key   TEXT,            -- minio object key  (storage = 'minio')
    data         BYTEA,           -- raw bytes          (storage = 'postgres')
    url          TEXT,            -- external URL       (storage = 'external')
    content_type VARCHAR(100),
    created_at   TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT images_storage_payload_check CHECK (
        (storage = 'minio'    AND object_key IS NOT NULL AND data IS NULL AND url IS NULL) OR
        (storage = 'postgres' AND data       IS NOT NULL AND object_key IS NULL AND url IS NULL) OR
        (storage = 'external' AND url        IS NOT NULL AND object_key IS NULL AND data IS NULL)
    )
);

ALTER TABLE services ADD COLUMN picture_id UUID REFERENCES images(id) ON DELETE SET NULL;

-- Single owner (admin) per salon.
CREATE UNIQUE INDEX uq_business_owner ON business_users (business_id) WHERE role = 'admin';

-- Soft-delete for business_users (employee removal keeps history).
ALTER TABLE business_users ADD COLUMN active BOOLEAN NOT NULL DEFAULT true;

-- Fast lookup of active employees per salon (availability queries filter active=true).
CREATE INDEX idx_business_users_active ON business_users (business_id) WHERE active;

-- Employee <-> Service assignment (capability, managed by admins).
CREATE TABLE employee_services (
    business_user_id UUID NOT NULL REFERENCES business_users(id) ON DELETE CASCADE,
    service_id       UUID NOT NULL REFERENCES services(id)       ON DELETE CASCADE,
    PRIMARY KEY (business_user_id, service_id)
);

-- Fast lookup of "which employees offer this service".
CREATE INDEX idx_employee_services_service ON employee_services (service_id);

-- Employee unavailability / vacation / blocked time.
CREATE TABLE employee_unavailability (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_user_id UUID NOT NULL REFERENCES business_users(id) ON DELETE CASCADE,
    start_time       TIMESTAMPTZ NOT NULL,
    end_time         TIMESTAMPTZ NOT NULL,
    reason           VARCHAR(500),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_employee_unavailability
    ON employee_unavailability (business_user_id, start_time, end_time);

-- No overlapping unavailability rows for the same employee (hard invariant).
ALTER TABLE employee_unavailability
    ADD CONSTRAINT employee_unavailability_no_overlap
    EXCLUDE USING gist (
        business_user_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    );

-- Appointments: extend status set, add created_by + cancellation_reason.
ALTER TABLE appointments DROP CONSTRAINT appointments_status_check;
ALTER TABLE appointments ADD CONSTRAINT appointments_status_check
    CHECK (status IN ('pending','confirmed','completed','cancelled','no_show'));

ALTER TABLE appointments
    ADD COLUMN created_by          VARCHAR(255),
    ADD COLUMN cancellation_reason VARCHAR(500);

UPDATE appointments SET created_by = COALESCE(created_by, customer_user_id);
ALTER TABLE appointments ALTER COLUMN created_by SET NOT NULL;

-- "Max 1 active booking per customer per salon per UTC day" (expression index;
-- no redundant column, so start_time can never drift from it).
CREATE UNIQUE INDEX uq_appointments_customer_day
    ON appointments (business_id, customer_user_id, ((start_time AT TIME ZONE 'UTC')::date))
    WHERE status <> 'cancelled';

-- Enforce that the booked employee offers the service. NOT VALID grandfathers
-- existing rows; every new/updated row is checked atomically.
ALTER TABLE appointments
    ADD CONSTRAINT fk_appointments_employee_service
    FOREIGN KEY (business_user_id, service_id)
    REFERENCES employee_services (business_user_id, service_id)
    NOT VALID;

-- Concurrency-safe slot booking: no overlapping non-cancelled/no-show
-- appointments for the same employee (hard invariant).
ALTER TABLE appointments
    ADD CONSTRAINT appointments_no_overlap_per_employee
    EXCLUDE USING gist (
        business_user_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    ) WHERE (status NOT IN ('cancelled', 'no_show'));
