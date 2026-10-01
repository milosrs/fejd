DROP INDEX IF EXISTS uq_services_slug_per_business;
ALTER TABLE services DROP COLUMN IF EXISTS slug;
