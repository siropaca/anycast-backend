-- キャラクターにアバター画像カラムを追加
ALTER TABLE characters ADD COLUMN avatar_id UUID REFERENCES images (id) ON DELETE SET NULL;

-- インデックス追加
CREATE INDEX idx_characters_avatar_id ON characters (avatar_id);
