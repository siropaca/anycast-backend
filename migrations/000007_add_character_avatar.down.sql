-- インデックス削除
DROP INDEX IF EXISTS idx_characters_avatar_id;

-- キャラクターのアバター画像カラムを削除
ALTER TABLE characters DROP COLUMN IF EXISTS avatar_id;
