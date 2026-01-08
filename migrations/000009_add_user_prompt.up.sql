-- users テーブルに user_prompt カラムを追加
ALTER TABLE users ADD COLUMN user_prompt text NOT NULL DEFAULT '';
