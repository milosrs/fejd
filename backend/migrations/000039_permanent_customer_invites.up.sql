-- Customer invites are permanent (print-and-stick QR codes): a NULL expires_at
-- means the invitation never expires. Employee/owner/realm-admin invites keep
-- their time-boxed expiry.
ALTER TABLE invitations ALTER COLUMN expires_at DROP NOT NULL;
