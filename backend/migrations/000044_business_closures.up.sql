-- Salon-level non-working days: whole days the salon is closed (e.g. holidays).
-- Marking a day here removes every provider's availability for that UTC date.

CREATE TABLE business_closures (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    closure_date DATE NOT NULL,
    reason       VARCHAR(500),
    UNIQUE(business_id, closure_date)
);

CREATE INDEX idx_business_closures_business_date ON business_closures(business_id, closure_date);
