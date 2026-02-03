-- Ride service: rides table (shared DB, users from auth)
CREATE TABLE IF NOT EXISTS rides (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id UUID NOT NULL,
    driver_id    UUID,
    status       TEXT NOT NULL CHECK (status IN ('requested', 'bidding', 'matched', 'in_progress', 'completed', 'cancelled')),
    from_lat     DOUBLE PRECISION NOT NULL,
    from_lng     DOUBLE PRECISION NOT NULL,
    from_address TEXT,
    to_lat       DOUBLE PRECISION NOT NULL,
    to_lng       DOUBLE PRECISION NOT NULL,
    to_address   TEXT,
    price        DOUBLE PRECISION,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rides_passenger ON rides (passenger_id);
CREATE INDEX IF NOT EXISTS idx_rides_driver ON rides (driver_id);
CREATE INDEX IF NOT EXISTS idx_rides_status ON rides (status);
CREATE INDEX IF NOT EXISTS idx_rides_created ON rides (created_at DESC);
