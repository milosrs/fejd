-- Restore the composite FK (NOT VALID to avoid validating historical rows that
-- may have been written after the constraint was dropped).
ALTER TABLE appointments
    ADD CONSTRAINT fk_appointments_employee_service
    FOREIGN KEY (business_user_id, service_id)
    REFERENCES employee_services (business_user_id, service_id)
    NOT VALID;
