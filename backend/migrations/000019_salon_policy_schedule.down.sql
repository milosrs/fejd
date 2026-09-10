DROP TABLE IF EXISTS business_hours;

ALTER TABLE businesses DROP COLUMN IF EXISTS slot_interval_minutes;
