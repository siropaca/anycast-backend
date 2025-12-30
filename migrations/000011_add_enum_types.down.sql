-- oauth_accounts.provider を VARCHAR に戻す
ALTER TABLE oauth_accounts
    ALTER COLUMN provider TYPE VARCHAR(50) USING provider::TEXT;

-- script_lines.line_type を VARCHAR に戻す
ALTER TABLE script_lines
    ALTER COLUMN line_type TYPE VARCHAR(50) USING line_type::TEXT;

-- voices.gender を VARCHAR に戻す
ALTER TABLE voices
    ALTER COLUMN gender TYPE VARCHAR(20) USING gender::TEXT;

-- enum 型を削除
DROP TYPE oauth_provider;
DROP TYPE line_type;
DROP TYPE gender;
