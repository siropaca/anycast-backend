-- episodes.user_prompt を NOT NULL に変更（デフォルト: 空文字）
UPDATE episodes SET user_prompt = '' WHERE user_prompt IS NULL;
ALTER TABLE episodes ALTER COLUMN user_prompt SET NOT NULL;
ALTER TABLE episodes ALTER COLUMN user_prompt SET DEFAULT '';
