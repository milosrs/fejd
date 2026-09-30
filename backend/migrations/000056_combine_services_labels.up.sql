-- Labels for combining multiple services into a single reservation.

INSERT INTO translations (key, locale, value) VALUES
    ('booking.addons.title',         'en', 'Add-on services'),
    ('booking.addons.title',         'rs', 'Dodatne usluge'),
    ('booking.addons.description',   'en', 'Combine additional services with your booking. Duration and price add up.'),
    ('booking.addons.description',   'rs', 'Kombinujte dodatne usluge uz rezervaciju. Trajanje i cena se sabiraju.'),
    ('booking.error.noCombination',  'en', 'This service combination is not available'),
    ('booking.error.noCombination',  'rs', 'Ova kombinacija usluga nije dostupna'),
    ('services.fields.combine',      'en', 'Can be combined with'),
    ('services.fields.combine',      'rs', 'Može se kombinovati sa'),
    ('services.noCombinations',      'en', 'No other services to combine with yet.'),
    ('services.noCombinations',      'rs', 'Još nema drugih usluga za kombinovanje.');
