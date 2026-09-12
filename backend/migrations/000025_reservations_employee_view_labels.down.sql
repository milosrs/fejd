DELETE FROM translations WHERE key IN (
    'reservations.myCalendar',
    'reservations.selectEmployee',
    'reservations.reserved',
    'reservations.reservedRequest',
    'reservations.reservedApproved',
    'reservations.reservedDenied'
);
