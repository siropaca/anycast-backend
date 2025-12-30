-- channels テーブルに公開日時カラムを追加
ALTER TABLE channels ADD COLUMN published_at TIMESTAMP;
CREATE INDEX idx_channels_published_at ON channels(published_at);

-- episodes テーブルに公開日時カラムを追加
ALTER TABLE episodes ADD COLUMN published_at TIMESTAMP;
CREATE INDEX idx_episodes_published_at ON episodes(published_at);
