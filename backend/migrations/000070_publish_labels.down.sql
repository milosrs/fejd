DELETE FROM translations WHERE key IN (
    'publish.title',
    'publish.help',
    'publish.button',
    'publish.success',
    'publish.noChanges',
    'publish.error'
);
