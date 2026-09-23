DELETE FROM translations WHERE key IN (
    'admin.inviteOwner',
    'admin.inviteCustomer',
    'admin.inviteRealmAdmin',
    'admin.inviteError',
    'invite.helpOwner',
    'invite.helpRealmAdmin',
    'invite.landing.invitedToPlatform'
);
