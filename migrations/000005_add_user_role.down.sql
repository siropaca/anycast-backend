-- users テーブルから role カラムを削除
ALTER TABLE users DROP COLUMN role;

-- user_role Enum 型を削除
DROP TYPE user_role;
