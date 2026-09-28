-- "Opening hours" landing-page section: allow the "hours" section type and add
-- its UI labels.

ALTER TABLE sections DROP CONSTRAINT sections_type_check;
ALTER TABLE sections ADD CONSTRAINT sections_type_check
    CHECK (type IN ('hero', 'about', 'gallery', 'contact', 'hours'));

INSERT INTO translations (key, locale, value) VALUES
    ('sections.hours',             'en', 'Opening hours'),
    ('sections.hours',             'rs', 'Radno vreme'),
    ('sections.hours.closed',      'en', 'Closed'),
    ('sections.hours.closed',      'rs', 'Zatvoreno'),
    ('sections.hours.openNow',     'en', 'Open now'),
    ('sections.hours.openNow',     'rs', 'Otvoreno'),
    ('sections.hours.closedNow',   'en', 'Closed now'),
    ('sections.hours.closedNow',   'rs', 'Zatvoreno');
