-- ============================================
-- Follows Table
-- ============================================

-- follows
CREATE TABLE follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_follows_user_id ON follows (user_id);
CREATE INDEX idx_follows_episode_id ON follows (episode_id);
