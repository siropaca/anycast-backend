-- channels.category_id を必須に変更
-- 既存データがある場合は先にカテゴリを設定する必要がある

-- 外部キー制約を再作成（ON DELETE SET NULL → ON DELETE RESTRICT）
ALTER TABLE channels
DROP CONSTRAINT IF EXISTS channels_category_id_fkey;

ALTER TABLE channels
ADD CONSTRAINT channels_category_id_fkey
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;

-- NOT NULL 制約を追加
ALTER TABLE channels
ALTER COLUMN category_id SET NOT NULL;
