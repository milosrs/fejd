-- Custom notification content for the appointment reminder plus a toggle for
-- the owner/staff "new appointment" notifications.

ALTER TABLE businesses
    ADD COLUMN staff_notifications_enabled BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN reminder_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN reminder_body TEXT NOT NULL DEFAULT '';

INSERT INTO translations (key, locale, value) VALUES
    ('policy.staffNotifications',     'en', 'New appointment notifications'),
    ('policy.staffNotifications',     'rs', 'Obaveštenja o novim terminima'),
    ('policy.staffNotificationsHelp', 'en', 'Notify me when a customer books an appointment with me.'),
    ('policy.staffNotificationsHelp', 'rs', 'Obavesti me kada klijent zakaže termin kod mene.'),
    ('policy.reminderTitle',          'en', 'Reminder title'),
    ('policy.reminderTitle',          'rs', 'Naslov podsetnika'),
    ('policy.reminderBody',           'en', 'Reminder message'),
    ('policy.reminderBody',           'rs', 'Tekst podsetnika'),
    ('policy.reminderBodyHelp',       'en', 'Shown to the customer before their appointment starts. Leave empty for the default.'),
    ('policy.reminderBodyHelp',       'rs', 'Prikazuje se klijentu pre početka termina. Ostavite prazno za podrazumevano.');
