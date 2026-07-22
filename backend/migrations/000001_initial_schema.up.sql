CREATE TABLE businesses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE business_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'employee')),
    display_name VARCHAR(255),
    UNIQUE(business_id, user_id)
);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    price DECIMAL(10,2),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE working_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_user_id UUID NOT NULL REFERENCES business_users(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL CHECK (end_time > start_time),
    UNIQUE(business_user_id, day_of_week)
);

CREATE TABLE working_hours_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_user_id UUID NOT NULL REFERENCES business_users(id) ON DELETE CASCADE,
    override_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    is_off BOOLEAN DEFAULT false,
    reason VARCHAR(500),
    UNIQUE(business_user_id, override_date)
);

CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    service_id UUID NOT NULL REFERENCES services(id),
    business_user_id UUID NOT NULL REFERENCES business_users(id),
    customer_user_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL CHECK (end_time > start_time),
    status VARCHAR(20) DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled', 'completed', 'no_show')),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_business_users_business ON business_users(business_id);
CREATE INDEX idx_business_users_user ON business_users(user_id);
CREATE INDEX idx_services_business ON services(business_id);
CREATE INDEX idx_appointments_business_time ON appointments(business_id, business_user_id, start_time);
CREATE INDEX idx_appointments_customer ON appointments(customer_user_id);
CREATE INDEX idx_working_hours_override_date ON working_hours_overrides(business_user_id, override_date);
