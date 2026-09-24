DELETE FROM translations WHERE key IN (
    'appointments.duration',
    'appointments.status.pending',
    'appointments.status.confirmed',
    'appointments.status.completed',
    'appointments.status.cancelled',
    'appointments.status.noShow'
);
