-- enum 型を作成
CREATE TYPE oauth_provider AS ENUM ('google');
CREATE TYPE line_type AS ENUM ('speech', 'silence', 'sfx');
CREATE TYPE gender AS ENUM ('male', 'female', 'neutral');

-- oauth_accounts.provider を enum に変更
ALTER TABLE oauth_accounts
    ALTER COLUMN provider TYPE oauth_provider USING provider::oauth_provider;

-- script_lines.line_type を enum に変更
ALTER TABLE script_lines
    ALTER COLUMN line_type TYPE line_type USING line_type::line_type;

-- voices.gender を enum に変更
ALTER TABLE voices
    ALTER COLUMN gender TYPE gender USING gender::gender;
