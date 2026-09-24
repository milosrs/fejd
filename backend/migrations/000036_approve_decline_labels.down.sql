UPDATE translations SET value = 'Accept'   WHERE key = 'reservations.accept' AND locale = 'en';
UPDATE translations SET value = 'Prihvati' WHERE key = 'reservations.accept' AND locale = 'rs';
UPDATE translations SET value = 'Reject'   WHERE key = 'reservations.reject' AND locale = 'en';
UPDATE translations SET value = 'Odbij'    WHERE key = 'reservations.reject' AND locale = 'rs';

DELETE FROM translations WHERE key IN (
    'policy.autoApprove',
    'policy.autoApproveHelp',
    'reservations.reasonOptional',
    'reservations.customerDetails',
    'reservations.serviceDetails',
    'reservations.time'
);
