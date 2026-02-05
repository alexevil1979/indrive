-- Payment providers: extend payments table, add payment_methods

-- Add new columns to payments
ALTER TABLE payments ADD COLUMN IF NOT EXISTS user_id UUID;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS provider TEXT DEFAULT 'cash';
ALTER TABLE payments ADD COLUMN IF NOT EXISTS confirm_url TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS fail_reason TEXT;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS refunded_at TIMESTAMPTZ;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ;

-- Update status constraint to include new statuses
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_status_check;
ALTER TABLE payments ADD CONSTRAINT payments_status_check 
    CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded'));

-- Add provider constraint
ALTER TABLE payments ADD CONSTRAINT payments_provider_check 
    CHECK (provider IN ('cash', 'tinkoff', 'yoomoney', 'sber'));

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE INDEX IF NOT EXISTS idx_payments_provider ON payments (provider);
CREATE INDEX IF NOT EXISTS idx_payments_external_id ON payments (external_id);

-- Saved payment methods (cards)
CREATE TABLE IF NOT EXISTS payment_methods (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL,
    provider     TEXT NOT NULL CHECK (provider IN ('tinkoff', 'yoomoney', 'sber')),
    type         TEXT NOT NULL DEFAULT 'card',
    last4        TEXT,
    brand        TEXT, -- visa, mastercard, mir
    expiry_month INT,
    expiry_year  INT,
    token_id     TEXT NOT NULL, -- Provider's token for recurring
    is_default   BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    UNIQUE(user_id, provider, token_id)
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_user ON payment_methods (user_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_default ON payment_methods (user_id) WHERE is_default = TRUE;

-- Refunds table
CREATE TABLE IF NOT EXISTS refunds (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id   UUID NOT NULL REFERENCES payments(id),
    amount       DOUBLE PRECISION NOT NULL CHECK (amount > 0),
    reason       TEXT,
    external_id  TEXT,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'succeeded', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds (payment_id);
