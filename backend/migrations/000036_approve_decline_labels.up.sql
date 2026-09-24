-- Approve/Decline button labels and the salon "automatic approval" policy.

UPDATE translations SET value = 'Approve' WHERE key = 'reservations.accept' AND locale = 'en';
UPDATE translations SET value = 'Odobri'  WHERE key = 'reservations.accept' AND locale = 'rs';
UPDATE translations SET value = 'Decline' WHERE key = 'reservations.reject' AND locale = 'en';
UPDATE translations SET value = 'Odbij'   WHERE key = 'reservations.reject' AND locale = 'rs';

INSERT INTO translations (key, locale, value) VALUES
    ('policy.autoApprove',        'en', 'Automatic appointment approval'),
    ('policy.autoApprove',        'rs', 'Automatsko odobravanje termina'),
    ('policy.autoApproveHelp',    'en', 'Approve new appointments automatically. Recommended for salons with many users that cannot keep track of all appointments.'),
    ('policy.autoApproveHelp',    'rs', 'Automatski odobri nove zakazane termine. Preporučeno za salone koji imaju puno korisnika i ne mogu da prate sve termine.'),
    ('reservations.reasonOptional',  'en', 'Reason (optional)'),
    ('reservations.reasonOptional',  'rs', 'Razlog (opciono)'),
    ('reservations.customerDetails', 'en', 'Customer details'),
    ('reservations.customerDetails', 'rs', 'Podaci o klijentu'),
    ('reservations.serviceDetails',  'en', 'Service details'),
    ('reservations.serviceDetails',  'rs', 'Podaci o usluzi'),
    ('reservations.time',            'en', 'Time'),
    ('reservations.time',            'rs', 'Vreme');
