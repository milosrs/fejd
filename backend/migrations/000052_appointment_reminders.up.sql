-- Appointment reminder policy (toggle + lead time) on the salon, and a
-- per-appointment stamp so the reminder job never sends twice.

ALTER TABLE businesses
    ADD COLUMN appointment_reminder_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN appointment_reminder_lead_minutes INT NOT NULL DEFAULT 60;

ALTER TABLE appointments
    ADD COLUMN reminder_sent_at TIMESTAMPTZ;

INSERT INTO translations (key, locale, value) VALUES
    ('policy.appointmentReminder',     'en', 'Appointment reminder'),
    ('policy.appointmentReminder',     'rs', 'Podsetnik za termin'),
    ('policy.appointmentReminderHelp', 'en', 'Send the customer a notification before their appointment starts.'),
    ('policy.appointmentReminderHelp', 'rs', 'Pošalji klijentu obaveštenje pre početka termina.'),
    ('policy.appointmentReminderLead', 'en', 'Remind before'),
    ('policy.appointmentReminderLead', 'rs', 'Podseti pre');
