DELETE FROM translations WHERE key IN (
    'nav.invitedCustomers',
    'invitedCustomers.title',
    'invitedCustomers.ownRegistration',
    'invitedCustomers.empty',
    'invitedCustomers.notAuthorized'
);
