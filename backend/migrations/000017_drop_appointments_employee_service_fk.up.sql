-- Unmapping an employee from a service must cancel their future reservations
-- rather than fail, and those cancelled rows must remain as history while the
-- employee_services mapping is removed. The composite FK blocked that (it
-- forbids deleting a mapping still referenced by any appointment, even a
-- cancelled one). The "employee offers the service" rule is already enforced in
-- the service layer at booking time, so the FK is dropped in favour of that
-- soft check.
ALTER TABLE appointments DROP CONSTRAINT IF EXISTS fk_appointments_employee_service;
