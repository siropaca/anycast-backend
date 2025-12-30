-- episodes テーブルから公開日時カラムを削除
DROP INDEX IF EXISTS idx_episodes_published_at;
ALTER TABLE episodes DROP COLUMN IF EXISTS published_at;

-- channels テーブルから公開日時カラムを削除
DROP INDEX IF EXISTS idx_channels_published_at;
ALTER TABLE channels DROP COLUMN IF EXISTS published_at;
