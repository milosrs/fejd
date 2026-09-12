-- Employee blocked-time reservations now require owner acknowledgment.
-- Existing rows are treated as already acknowledged (confirmed).

ALTER TABLE employee_unavailability
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'confirmed';
