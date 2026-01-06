-- ユーザーロールの Enum 型を作成
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- users テーブルに role カラムを追加
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'user';
