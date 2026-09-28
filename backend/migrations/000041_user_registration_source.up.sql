-- Track how each user registered and, for customers, which salon invited them.
-- registration_source defaults to 'self'; the invite flow flips it to 'invite'.
-- invited_business_id is NULL for platform ("invite a friend") invites.
ALTER TABLE users ADD COLUMN registration_source TEXT NOT NULL DEFAULT 'self';
ALTER TABLE users ADD COLUMN invited_business_id UUID REFERENCES businesses(id) ON DELETE SET NULL;
