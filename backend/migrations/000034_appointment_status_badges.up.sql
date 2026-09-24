-- Appointment status badge labels (customer "my appointments" list).

INSERT INTO translations (key, locale, value) VALUES
    ('appointments.duration',           'en', 'Duration: {duration} min'),
    ('appointments.duration',           'rs', 'Trajanje: {duration} min'),
    ('appointments.status.pending',     'en', 'Pending approval'),
    ('appointments.status.pending',     'rs', 'Na čekanju'),
    ('appointments.status.confirmed',   'en', 'Approved'),
    ('appointments.status.confirmed',   'rs', 'Odobreno'),
    ('appointments.status.completed',   'en', 'Completed'),
    ('appointments.status.completed',   'rs', 'Završeno'),
    ('appointments.status.cancelled',   'en', 'Cancelled'),
    ('appointments.status.cancelled',   'rs', 'Otkazano'),
    ('appointments.status.noShow',      'en', 'No-show'),
    ('appointments.status.noShow',      'rs', 'Nepojavljivanje');
