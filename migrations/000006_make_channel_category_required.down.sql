-- channels.category_id を任意に戻す

-- NOT NULL 制約を削除
ALTER TABLE channels
ALTER COLUMN category_id DROP NOT NULL;

-- 外部キー制約を再作成（ON DELETE RESTRICT → ON DELETE SET NULL）
ALTER TABLE channels
DROP CONSTRAINT IF EXISTS channels_category_id_fkey;

ALTER TABLE channels
ADD CONSTRAINT channels_category_id_fkey
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
