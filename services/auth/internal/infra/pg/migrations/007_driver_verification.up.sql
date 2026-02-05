-- Driver verification: documents and verification requests
-- Extends profiles with full verification flow

-- Driver verification requests
CREATE TABLE IF NOT EXISTS driver_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    license_number VARCHAR(50),
    vehicle_model VARCHAR(100),
    vehicle_plate VARCHAR(20),
    vehicle_year INT,
    reject_reason TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(user_id) -- One active verification per user
);

-- Driver documents (uploaded files)
CREATE TABLE IF NOT EXISTS driver_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    verification_id UUID REFERENCES driver_verifications(id) ON DELETE SET NULL,
    doc_type VARCHAR(30) NOT NULL, -- license, passport, vehicle_reg, insurance, photo, vehicle_photo
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100),
    storage_key VARCHAR(500) NOT NULL, -- S3/MinIO key
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    reject_reason TEXT,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_driver_verifications_user ON driver_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_driver_verifications_status ON driver_verifications(status);
CREATE INDEX IF NOT EXISTS idx_driver_documents_user ON driver_documents(user_id);
CREATE INDEX IF NOT EXISTS idx_driver_documents_verification ON driver_documents(verification_id);
CREATE INDEX IF NOT EXISTS idx_driver_documents_type ON driver_documents(user_id, doc_type);

-- Update profiles to link to verification
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS verification_id UUID REFERENCES driver_verifications(id);
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT FALSE;
