-- script_prompt を NULLABLE に戻す

ALTER TABLE channels ALTER COLUMN script_prompt DROP NOT NULL;
ALTER TABLE episodes ALTER COLUMN script_prompt DROP NOT NULL;
