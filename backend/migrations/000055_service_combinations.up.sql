CREATE TABLE service_combinations (
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    combinable_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (service_id, combinable_service_id),
    CHECK (service_id <> combinable_service_id)
);

CREATE INDEX idx_service_combinations_service ON service_combinations(service_id);

CREATE TABLE appointment_services (
    appointment_id UUID NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id),
    position INT NOT NULL DEFAULT 0,
    PRIMARY KEY (appointment_id, service_id)
);

CREATE INDEX idx_appointment_services_service ON appointment_services(service_id);
