-- Salon-level non-working days. A closure is a rule that makes the salon
-- unavailable on matching days. Four rule kinds are supported:
--   single  - one specific date (start_date)
--   range   - a contiguous span of dates (start_date .. end_date, inclusive)
--   weekly  - a recurring day of week (day_of_week 0=Sun .. 6=Sat)
--   yearly  - a date that recurs every year (month + day)
-- The CHECK constraints enforce that each rule stores exactly the fields its
-- kind needs and nothing else.

CREATE TABLE business_closures (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id  UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    closure_type TEXT NOT NULL CHECK (closure_type IN ('single', 'range', 'weekly', 'yearly')),
    start_date   DATE,
    end_date     DATE,
    day_of_week  SMALLINT CHECK (day_of_week BETWEEN 0 AND 6),
    month        SMALLINT CHECK (month BETWEEN 1 AND 12),
    day          SMALLINT CHECK (day BETWEEN 1 AND 31),
    reason       VARCHAR(500),
    CHECK (
        (closure_type = 'single' AND start_date IS NOT NULL AND end_date IS NULL AND day_of_week IS NULL AND month IS NULL AND day IS NULL)
        OR
        (closure_type = 'range' AND start_date IS NOT NULL AND end_date IS NOT NULL AND end_date >= start_date AND day_of_week IS NULL AND month IS NULL AND day IS NULL)
        OR
        (closure_type = 'weekly' AND day_of_week IS NOT NULL AND start_date IS NULL AND end_date IS NULL AND month IS NULL AND day IS NULL)
        OR
        (closure_type = 'yearly' AND month IS NOT NULL AND day IS NOT NULL AND start_date IS NULL AND end_date IS NULL AND day_of_week IS NULL)
    )
);

CREATE INDEX idx_business_closures_business ON business_closures(business_id);
