ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_no_overlap_per_employee;
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_employee_service;
DROP INDEX IF EXISTS uq_appointments_customer_day;
ALTER TABLE appointments DROP COLUMN IF EXISTS cancellation_reason;
ALTER TABLE appointments DROP COLUMN IF EXISTS created_by;
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS appointments_status_check;
ALTER TABLE appointments ADD CONSTRAINT appointments_status_check
    CHECK (status IN ('confirmed','cancelled','completed','no_show'));
DROP TABLE IF EXISTS employee_unavailability;
DROP TABLE IF EXISTS employee_services;
DROP INDEX IF EXISTS idx_business_users_active;
ALTER TABLE business_users DROP COLUMN IF EXISTS active;
DROP INDEX IF EXISTS uq_business_owner;
ALTER TABLE services DROP COLUMN IF EXISTS picture_id;
DROP TABLE IF EXISTS images;
ALTER TABLE services DROP COLUMN IF EXISTS description;
