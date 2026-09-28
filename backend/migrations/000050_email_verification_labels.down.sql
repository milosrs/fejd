DELETE FROM translations WHERE key IN (
    'verifyEmail.title',
    'verifyEmail.body',
    'verifyEmail.disabled'
);
