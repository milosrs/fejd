UPDATE translations SET value = 'Salon policy' WHERE key = 'policy.title' AND locale = 'en';

DELETE FROM translations WHERE key IN ('policy.description', 'policy.appointmentsHelp');
