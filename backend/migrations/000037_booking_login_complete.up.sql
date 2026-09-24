-- Anonymous visitors build the full reservation on the UI and are only asked to
-- log in at the final confirmation. The old service-card note that said "you have
-- to register to book" no longer applies, so it is dropped here.

INSERT INTO translations (key, locale, value) VALUES
    ('booking.loginToComplete', 'en', 'Log in to complete the reservation'),
    ('booking.loginToComplete', 'rs', 'Prijavite se da dovršite rezervaciju');

DELETE FROM translations WHERE key = 'services.book.requiresAuth';
