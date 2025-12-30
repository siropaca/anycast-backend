-- ============================================
-- Playback Histories Table
-- ============================================

-- playback_histories
CREATE TABLE playback_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    progress_ms INTEGER NOT NULL DEFAULT 0,
    completed BOOLEAN NOT NULL DEFAULT false,
    played_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_playback_histories_user_id ON playback_histories (user_id);
CREATE INDEX idx_playback_histories_episode_id ON playback_histories (episode_id);
CREATE INDEX idx_playback_histories_user_id_played_at ON playback_histories (user_id, played_at);
