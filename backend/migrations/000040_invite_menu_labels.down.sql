DELETE FROM translations WHERE key IN (
    'invite.menu.section',
    'invite.menu.employee',
    'invite.menu.customer',
    'invite.menu.owner',
    'invite.menu.realmAdmin',
    'invite.permanent'
);

UPDATE translations SET value = 'No appointments yet.'
    WHERE key = 'appointments.empty' AND locale = 'en';
UPDATE translations SET value = 'Još nema termina.'
    WHERE key = 'appointments.empty' AND locale = 'rs';
