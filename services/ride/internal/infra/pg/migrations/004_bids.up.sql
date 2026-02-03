-- Ride service: bids table (inDrive-style price bidding)
CREATE TABLE IF NOT EXISTS bids (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id    UUID NOT NULL REFERENCES rides (id) ON DELETE CASCADE,
    driver_id  UUID NOT NULL,
    price      DOUBLE PRECISION NOT NULL CHECK (price > 0),
    status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_bids_ride ON bids (ride_id);
CREATE INDEX IF NOT EXISTS idx_bids_driver ON bids (driver_id);
