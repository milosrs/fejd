-- Follow-up: make the employee-service FK unconditional once pre-migration
-- data has been reconciled (see plan). Safe on a fresh database.
ALTER TABLE appointments VALIDATE CONSTRAINT fk_appointments_employee_service;
