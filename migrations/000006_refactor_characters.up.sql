-- Character を Channel から User の所有に変更するマイグレーション

-- 1. channel_characters 中間テーブルを作成
CREATE TABLE channel_characters (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	channel_id UUID NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
	character_id UUID NOT NULL REFERENCES characters (id) ON DELETE RESTRICT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (channel_id, character_id)
);

CREATE INDEX idx_channel_characters_channel_id ON channel_characters (channel_id);
CREATE INDEX idx_channel_characters_character_id ON channel_characters (character_id);

-- 2. 既存データを channel_characters にマイグレーション
INSERT INTO channel_characters (channel_id, character_id)
SELECT channel_id, id FROM characters;

-- 3. characters テーブルに user_id カラムを追加
ALTER TABLE characters ADD COLUMN user_id UUID;

-- 4. 既存データの user_id を設定（channel の user_id を取得）
UPDATE characters
SET user_id = channels.user_id
FROM channels
WHERE characters.channel_id = channels.id;

-- 5. user_id を NOT NULL に変更
ALTER TABLE characters ALTER COLUMN user_id SET NOT NULL;

-- 6. 外部キー制約を追加
ALTER TABLE characters
ADD CONSTRAINT fk_characters_user_id
FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- 7. 古いインデックスと制約を削除
DROP INDEX idx_characters_channel_id;
ALTER TABLE characters DROP CONSTRAINT characters_channel_id_name_key;

-- 8. channel_id カラムを削除
ALTER TABLE characters DROP COLUMN channel_id;

-- 9. 新しいインデックスと制約を追加（user_id + name で一意）
CREATE INDEX idx_characters_user_id ON characters (user_id);
ALTER TABLE characters ADD CONSTRAINT characters_user_id_name_key UNIQUE (user_id, name);
