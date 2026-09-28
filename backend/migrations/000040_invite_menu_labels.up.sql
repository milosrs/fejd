-- Invite options moved into the profile dropdown menu, plus the appointment
-- empty-state copy and the "never expires" label for permanent customer QR
-- invites.

INSERT INTO translations (key, locale, value) VALUES
    ('invite.menu.section',    'en', 'Invite'),
    ('invite.menu.section',    'rs', 'Pozovi'),
    ('invite.menu.employee',   'en', 'Invite employee'),
    ('invite.menu.employee',   'rs', 'Pozovi zaposlenog'),
    ('invite.menu.customer',   'en', 'Invite customer'),
    ('invite.menu.customer',   'rs', 'Pozovi mušteriju'),
    ('invite.menu.owner',      'en', 'Invite owner'),
    ('invite.menu.owner',      'rs', 'Pozovi vlasnika'),
    ('invite.menu.realmAdmin', 'en', 'Invite realm admin'),
    ('invite.menu.realmAdmin', 'rs', 'Pozovi realm administratora'),
    ('invite.permanent',       'en', 'Never expires'),
    ('invite.permanent',       'rs', 'Ne ističe');

UPDATE translations SET value = 'There are no appointments created for this account.'
    WHERE key = 'appointments.empty' AND locale = 'en';
UPDATE translations SET value = 'Nema termina kreiranih za ovaj nalog.'
    WHERE key = 'appointments.empty' AND locale = 'rs';
