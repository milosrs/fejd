-- User profile picture. The image lives in the generic `images` table (no
-- business_id, since a profile picture belongs to the user, not a salon) and is
-- referenced directly here. ON DELETE SET NULL keeps the users row intact if an
-- avatar image is ever removed out-of-band.
ALTER TABLE users ADD COLUMN avatar_id UUID REFERENCES images(id) ON DELETE SET NULL;
