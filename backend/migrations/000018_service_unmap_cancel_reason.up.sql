-- Cancellation reason label shown when a service is unmapped from an employee
-- and their future reservations are cancelled. {date} is interpolated on the
-- frontend with the unmapping date.
INSERT INTO translations (key, locale, value) VALUES
    ('cancellation.reason.noService', 'en', 'Does not offer this service as of {date}'),
    ('cancellation.reason.noService', 'rs', 'Ne nudi ovu uslugu od {date}');
