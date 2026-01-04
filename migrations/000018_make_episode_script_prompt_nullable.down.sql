-- episodes テーブルの script_prompt を NOT NULL に戻す
-- 既存の NULL データを空文字に変換
UPDATE episodes SET script_prompt = '' WHERE script_prompt IS NULL;

ALTER TABLE episodes ALTER COLUMN script_prompt SET NOT NULL;
