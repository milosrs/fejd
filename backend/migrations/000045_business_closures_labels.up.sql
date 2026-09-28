-- Salon policy "non-working days" labels.

INSERT INTO translations (key, locale, value) VALUES
    ('policy.nonWorkingDays',      'en', 'Non-working days'),
    ('policy.nonWorkingDays',      'rs', 'Neradni dani'),
    ('policy.nonWorkingDaysHelp',  'en', 'Mark whole days the salon is closed. No appointments are offered on these days.'),
    ('policy.nonWorkingDaysHelp',  'rs', 'Označi cele dane kada je salon zatvoren. Tih dana se ne nude termini.'),
    ('policy.nonWorkingDayDate',   'en', 'Date'),
    ('policy.nonWorkingDayDate',   'rs', 'Datum'),
    ('policy.nonWorkingDayReason', 'en', 'Reason (optional)'),
    ('policy.nonWorkingDayReason', 'rs', 'Razlog (opciono)'),
    ('policy.addNonWorkingDay',    'en', 'Add day'),
    ('policy.addNonWorkingDay',    'rs', 'Dodaj dan'),
    ('policy.nonWorkingDayInvalid','en', 'Enter a valid date (YYYY-MM-DD).'),
    ('policy.nonWorkingDayInvalid','rs', 'Unesite važeći datum (GGGG-MM-DD).'),
    ('policy.nonWorkingDaysEmpty', 'en', 'No non-working days yet.'),
    ('policy.nonWorkingDaysEmpty', 'rs', 'Još nema neradnih dana.');
