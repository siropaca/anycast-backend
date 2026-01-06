-- Character を User から Channel の所有に戻すマイグレーション

-- 1. characters テーブルに channel_id カラムを追加
ALTER TABLE characters ADD COLUMN channel_id UUID;

-- 2. channel_characters から channel_id を復元
-- 注意: 複数チャンネルに紐づいている場合は最初の1つを使用
UPDATE characters
SET channel_id = (
	SELECT channel_id
	FROM channel_characters
	WHERE channel_characters.character_id = characters.id
	LIMIT 1
);

-- 3. channel_id を NOT NULL に変更（紐づけがないキャラクターは削除）
DELETE FROM characters WHERE channel_id IS NULL;
ALTER TABLE characters ALTER COLUMN channel_id SET NOT NULL;

-- 4. 外部キー制約を追加
ALTER TABLE characters
ADD CONSTRAINT fk_characters_channel_id
FOREIGN KEY (channel_id) REFERENCES channels (id) ON DELETE CASCADE;

-- 5. 新しいインデックスと制約を削除
DROP INDEX idx_characters_user_id;
ALTER TABLE characters DROP CONSTRAINT characters_user_id_name_key;

-- 6. user_id の外部キー制約を削除
ALTER TABLE characters DROP CONSTRAINT fk_characters_user_id;

-- 7. user_id カラムを削除
ALTER TABLE characters DROP COLUMN user_id;

-- 8. 古いインデックスと制約を追加
CREATE INDEX idx_characters_channel_id ON characters (channel_id);
ALTER TABLE characters ADD CONSTRAINT characters_channel_id_name_key UNIQUE (channel_id, name);

-- 9. channel_characters テーブルを削除
DROP TABLE channel_characters;
