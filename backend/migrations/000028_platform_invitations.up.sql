ALTER TABLE invitations ALTER COLUMN business_id DROP NOT NULL;
ALTER TABLE invitations DROP CONSTRAINT invitations_role_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_role_check CHECK (role IN ('employee', 'customer', 'owner'));
