-- episodes テーブルの script_prompt を NULL 許可に変更
ALTER TABLE episodes ALTER COLUMN script_prompt DROP NOT NULL;
