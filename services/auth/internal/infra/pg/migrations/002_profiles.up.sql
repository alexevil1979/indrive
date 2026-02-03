-- User service: profiles + driver_profiles (shared DB)
CREATE TABLE IF NOT EXISTS profiles (
    user_id     UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    display_name TEXT,
    phone       TEXT,
    avatar_url  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS driver_profiles (
    user_id          UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    verified         BOOLEAN NOT NULL DEFAULT false,
    license_number   TEXT,
    doc_license_url  TEXT,
    doc_photo_url    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_profiles_phone ON profiles (phone);
