-- FCM device tokens for push notifications, keyed by the Keycloak subject so a
-- user can be targeted across devices (native + web). `token` is unique because
-- a single FCM registration token identifies exactly one device instance.

CREATE TABLE push_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    VARCHAR(255) NOT NULL,
    token      TEXT NOT NULL,
    platform   VARCHAR(20) NOT NULL DEFAULT 'unknown',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (token)
);

CREATE INDEX idx_push_tokens_user ON push_tokens(user_id);
