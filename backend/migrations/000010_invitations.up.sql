CREATE TABLE invitations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    role        TEXT NOT NULL DEFAULT 'employee' CHECK (role IN ('employee')),
    created_by  TEXT NOT NULL,
    max_uses    INTEGER NOT NULL DEFAULT 1 CHECK (max_uses > 0),
    use_count   INTEGER NOT NULL DEFAULT 0 CHECK (use_count >= 0),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invitations_business ON invitations (business_id);
