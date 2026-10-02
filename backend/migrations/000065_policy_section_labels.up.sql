-- Salon policy page: page-level title + description, and section descriptions.

UPDATE translations SET value = 'Salon Policy' WHERE key = 'policy.title' AND locale = 'en';

INSERT INTO translations (key, locale, value) VALUES
    ('policy.description',      'en', 'Manage your working time, days off, no-show policy and more'),
    ('policy.description',      'rs', 'Upravljajte radnim vremenom, slobodnim danima, politikom nepojavljivanja i još mnogo toga'),
    ('policy.appointmentsHelp', 'en', 'Control approvals, reminders, cancellation notice and time slots.'),
    ('policy.appointmentsHelp', 'rs', 'Kontrolišite odobravanja, podsetnike, otkazni rok i intervale termina.');
