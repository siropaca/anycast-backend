-- username カラムを削除し、display_name を name に戻す

-- username の一意制約を削除
DROP INDEX IF EXISTS idx_users_username;

-- username カラムを削除
ALTER TABLE users DROP COLUMN IF EXISTS username;

-- display_name を name に戻す
ALTER TABLE users RENAME COLUMN display_name TO name;
