-- Ratings table
CREATE TABLE IF NOT EXISTS ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('passenger', 'driver')),
    score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT,
    tags TEXT[], -- array of tag strings
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure one rating per ride per direction
    UNIQUE (ride_id, from_user_id, to_user_id)
);

-- Indexes for efficient queries
CREATE INDEX idx_ratings_to_user_id ON ratings(to_user_id);
CREATE INDEX idx_ratings_from_user_id ON ratings(from_user_id);
CREATE INDEX idx_ratings_ride_id ON ratings(ride_id);
CREATE INDEX idx_ratings_role ON ratings(role);
CREATE INDEX idx_ratings_created_at ON ratings(created_at DESC);

-- User ratings aggregate (materialized for performance)
CREATE TABLE IF NOT EXISTS user_ratings (
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('passenger', 'driver')),
    average_score DECIMAL(3,2) NOT NULL DEFAULT 0,
    total_ratings INTEGER NOT NULL DEFAULT 0,
    score_5_count INTEGER NOT NULL DEFAULT 0,
    score_4_count INTEGER NOT NULL DEFAULT 0,
    score_3_count INTEGER NOT NULL DEFAULT 0,
    score_2_count INTEGER NOT NULL DEFAULT 0,
    score_1_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (user_id, role)
);

-- Function to update user_ratings aggregate
CREATE OR REPLACE FUNCTION update_user_rating()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO user_ratings (user_id, role, average_score, total_ratings, 
        score_5_count, score_4_count, score_3_count, score_2_count, score_1_count, updated_at)
    SELECT 
        NEW.to_user_id,
        NEW.role,
        AVG(score)::DECIMAL(3,2),
        COUNT(*),
        COUNT(*) FILTER (WHERE score = 5),
        COUNT(*) FILTER (WHERE score = 4),
        COUNT(*) FILTER (WHERE score = 3),
        COUNT(*) FILTER (WHERE score = 2),
        COUNT(*) FILTER (WHERE score = 1),
        NOW()
    FROM ratings
    WHERE to_user_id = NEW.to_user_id AND role = NEW.role
    ON CONFLICT (user_id, role) DO UPDATE SET
        average_score = EXCLUDED.average_score,
        total_ratings = EXCLUDED.total_ratings,
        score_5_count = EXCLUDED.score_5_count,
        score_4_count = EXCLUDED.score_4_count,
        score_3_count = EXCLUDED.score_3_count,
        score_2_count = EXCLUDED.score_2_count,
        score_1_count = EXCLUDED.score_1_count,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update aggregates
CREATE TRIGGER trg_update_user_rating
    AFTER INSERT ON ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_user_rating();
