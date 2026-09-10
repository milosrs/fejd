-- Salon-level schedule configuration exposed through the salon policy: a
-- configurable booking slot interval and default weekly working hours that act
-- as a fallback for staff without their own hours.

ALTER TABLE businesses
    ADD COLUMN slot_interval_minutes INT NOT NULL DEFAULT 30;

CREATE TABLE business_hours (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    day_of_week  SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time   TIME NOT NULL,
    end_time     TIME NOT NULL CHECK (end_time > start_time),
    UNIQUE(business_id, day_of_week)
);

-- Seed default 09:00–17:00 hours for existing businesses.
INSERT INTO business_hours (business_id, day_of_week, start_time, end_time)
SELECT b.id, d.day_of_week, '09:00:00'::time, '17:00:00'::time
FROM businesses b
CROSS JOIN (VALUES (0), (1), (2), (3), (4), (5), (6)) AS d(day_of_week);
