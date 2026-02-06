-- Application settings table
-- Stores global configuration including map provider selection

CREATE TABLE IF NOT EXISTS app_settings (
    id VARCHAR(50) PRIMARY KEY DEFAULT 'default',
    map_provider VARCHAR(20) NOT NULL DEFAULT 'google',
    google_maps_api_key TEXT,
    yandex_maps_api_key TEXT,
    default_language VARCHAR(10) DEFAULT 'ru',
    default_currency VARCHAR(10) DEFAULT 'RUB',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    updated_by UUID REFERENCES users(id)
);

-- Insert default settings
INSERT INTO app_settings (id, map_provider, default_language, default_currency)
VALUES ('default', 'google', 'ru', 'RUB')
ON CONFLICT (id) DO NOTHING;

-- Settings history for audit
CREATE TABLE IF NOT EXISTS app_settings_history (
    id SERIAL PRIMARY KEY,
    settings_id VARCHAR(50) NOT NULL,
    map_provider VARCHAR(20),
    changed_at TIMESTAMPTZ DEFAULT NOW(),
    changed_by UUID,
    old_values JSONB,
    new_values JSONB
);

-- Trigger to track settings changes
CREATE OR REPLACE FUNCTION log_settings_change()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO app_settings_history (settings_id, map_provider, changed_by, old_values, new_values)
    VALUES (
        NEW.id,
        NEW.map_provider,
        NEW.updated_by,
        to_jsonb(OLD),
        to_jsonb(NEW)
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS settings_change_trigger ON app_settings;
CREATE TRIGGER settings_change_trigger
    AFTER UPDATE ON app_settings
    FOR EACH ROW
    EXECUTE FUNCTION log_settings_change();

-- Index for quick lookup
CREATE INDEX IF NOT EXISTS idx_settings_history_time ON app_settings_history(changed_at DESC);
