-- Mark salons created by a Keycloak realm administrator so the app can hide
-- them from the public directory and public salon routes for everyone except
-- realm administrators.
ALTER TABLE businesses
    ADD COLUMN realm_admin_created BOOLEAN NOT NULL DEFAULT false;
