-- ============================================
-- Likes Table
-- ============================================

-- likes
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_likes_user_id ON likes (user_id);
CREATE INDEX idx_likes_episode_id ON likes (episode_id);
