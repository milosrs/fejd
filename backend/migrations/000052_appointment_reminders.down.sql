ALTER TABLE appointments
    DROP COLUMN IF EXISTS reminder_sent_at;

ALTER TABLE businesses
    DROP COLUMN IF EXISTS appointment_reminder_lead_minutes,
    DROP COLUMN IF EXISTS appointment_reminder_enabled;

DELETE FROM translations WHERE key IN
    ('policy.appointmentReminder', 'policy.appointmentReminderHelp', 'policy.appointmentReminderLead');
