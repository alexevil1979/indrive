-- Payment service: payments table (checkout flow, cash/card stub)
CREATE TABLE IF NOT EXISTS payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id     UUID NOT NULL,
    amount      DOUBLE PRECISION NOT NULL CHECK (amount > 0),
    currency    TEXT NOT NULL DEFAULT 'RUB',
    method      TEXT NOT NULL CHECK (method IN ('cash', 'card')),
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    external_id TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_ride ON payments (ride_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status);
