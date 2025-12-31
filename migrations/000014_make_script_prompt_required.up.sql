-- script_prompt を NOT NULL に変更

-- channels: NULL の場合は空文字列で埋める
UPDATE channels SET script_prompt = '' WHERE script_prompt IS NULL;
ALTER TABLE channels ALTER COLUMN script_prompt SET NOT NULL;

-- episodes: NULL の場合は空文字列で埋める
UPDATE episodes SET script_prompt = '' WHERE script_prompt IS NULL;
ALTER TABLE episodes ALTER COLUMN script_prompt SET NOT NULL;
