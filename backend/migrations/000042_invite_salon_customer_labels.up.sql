-- Labels for the employee invite choice (friend vs salon customer) and the
-- platform-customer ("friend") invite help text.

INSERT INTO translations (key, locale, value) VALUES
    ('invite.menu.salonCustomer', 'en', 'Invite salon customer'),
    ('invite.menu.salonCustomer', 'rs', 'Pozovi mušteriju salona'),
    ('invite.helpFriend',         'en', 'Share this link or QR code. Scanning it lets someone join Fejd as a customer.'),
    ('invite.helpFriend',         'rs', 'Podelite ovaj link ili QR kod. Skeniranjem se neko može pridružiti Fejd-u kao mušterija.');
