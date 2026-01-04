-- user_prompt を script_prompt に戻す
ALTER TABLE channels RENAME COLUMN user_prompt TO script_prompt;
ALTER TABLE episodes RENAME COLUMN user_prompt TO script_prompt;
