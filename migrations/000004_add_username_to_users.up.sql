-- users テーブルに username カラムを追加し、name を display_name にリネーム

-- name を display_name にリネーム
ALTER TABLE users RENAME COLUMN name TO display_name;

-- username カラムを追加
ALTER TABLE users ADD COLUMN username VARCHAR(20) NOT NULL DEFAULT '';

-- 既存データの username を仮設定（id の先頭8文字を使用）
UPDATE users SET username = CONCAT('user_', LEFT(REPLACE(id::text, '-', ''), 8)) WHERE username = '';

-- デフォルト値を削除
ALTER TABLE users ALTER COLUMN username DROP DEFAULT;

-- username に一意制約を追加
CREATE UNIQUE INDEX idx_users_username ON users(username);
