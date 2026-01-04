-- script_prompt を user_prompt にリネーム
ALTER TABLE channels RENAME COLUMN script_prompt TO user_prompt;
ALTER TABLE episodes RENAME COLUMN script_prompt TO user_prompt;
