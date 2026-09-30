DELETE FROM translations WHERE key IN (
    'booking.addons.title',
    'booking.addons.description',
    'booking.error.noCombination',
    'services.fields.combine',
    'services.noCombinations'
);
