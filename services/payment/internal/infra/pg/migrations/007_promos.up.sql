-- Promo codes table
CREATE TABLE IF NOT EXISTS promos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('percent', 'fixed')),
    value DECIMAL(10,2) NOT NULL CHECK (value > 0),
    min_order_value DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_discount DECIMAL(10,2) NOT NULL DEFAULT 0,
    usage_limit INTEGER NOT NULL DEFAULT 0,
    usage_count INTEGER NOT NULL DEFAULT 0,
    per_user_limit INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for promos
CREATE INDEX idx_promos_code ON promos(code);
CREATE INDEX idx_promos_is_active ON promos(is_active);
CREATE INDEX idx_promos_expires_at ON promos(expires_at);

-- User promo usage tracking
CREATE TABLE IF NOT EXISTS user_promos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    ride_id UUID,
    discount DECIMAL(10,2) NOT NULL,
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint to track per-user usage
    UNIQUE (user_id, promo_id, ride_id)
);

-- Indexes for user_promos
CREATE INDEX idx_user_promos_user_id ON user_promos(user_id);
CREATE INDEX idx_user_promos_promo_id ON user_promos(promo_id);
CREATE INDEX idx_user_promos_ride_id ON user_promos(ride_id);

-- Function to increment promo usage count
CREATE OR REPLACE FUNCTION increment_promo_usage()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE promos SET 
        usage_count = usage_count + 1,
        updated_at = NOW()
    WHERE id = NEW.promo_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-increment usage
CREATE TRIGGER trg_increment_promo_usage
    AFTER INSERT ON user_promos
    FOR EACH ROW
    EXECUTE FUNCTION increment_promo_usage();

-- Insert some default promo codes for testing
INSERT INTO promos (code, description, type, value, min_order_value, max_discount, per_user_limit, starts_at) VALUES
    ('WELCOME10', 'Скидка 10% на первую поездку', 'percent', 10, 200, 500, 1, NOW()),
    ('SAVE100', 'Скидка 100₽ на любую поездку', 'fixed', 100, 300, 0, 0, NOW()),
    ('VIP20', 'VIP скидка 20%', 'percent', 20, 500, 1000, 0, NOW())
ON CONFLICT (code) DO NOTHING;
