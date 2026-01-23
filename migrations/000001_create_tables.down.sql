-- テーブルを依存関係の逆順で削除

DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS playback_histories;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS script_lines;
DROP TABLE IF EXISTS script_jobs;
DROP TABLE IF EXISTS audio_jobs;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS channel_characters;
DROP TABLE IF EXISTS bgms;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS system_bgms;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS voices;
DROP TABLE IF EXISTS audios;
DROP TABLE IF EXISTS images;

-- ENUM 型を削除
DROP TYPE IF EXISTS script_job_status;
DROP TYPE IF EXISTS audio_job_status;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS gender;
DROP TYPE IF EXISTS oauth_provider;
