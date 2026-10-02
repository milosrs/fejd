-- Rename the "My reservations" nav label to "My salon appointments".

UPDATE translations SET value = 'My salon appointments' WHERE key = 'nav.myReservations' AND locale = 'en';
UPDATE translations SET value = 'Moji salonski termini' WHERE key = 'nav.myReservations' AND locale = 'rs';
