DELETE FROM invitations WHERE business_id IS NULL;
ALTER TABLE invitations ALTER COLUMN business_id SET NOT NULL;
ALTER TABLE invitations DROP CONSTRAINT invitations_role_check;
ALTER TABLE invitations ADD CONSTRAINT invitations_role_check CHECK (role IN ('employee', 'customer'));
