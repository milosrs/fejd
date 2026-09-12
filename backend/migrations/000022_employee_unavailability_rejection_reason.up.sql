-- Track why the owner rejected an employee's blocked-time reservation.

ALTER TABLE employee_unavailability
    ADD COLUMN rejection_reason TEXT;
