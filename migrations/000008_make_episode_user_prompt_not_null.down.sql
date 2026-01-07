-- episodes.user_prompt を NULLABLE に戻す
ALTER TABLE episodes ALTER COLUMN user_prompt DROP DEFAULT;
ALTER TABLE episodes ALTER COLUMN user_prompt DROP NOT NULL;
