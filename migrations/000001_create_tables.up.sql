-- Anycast データベーススキーマ
-- 全テーブルを最終形の状態で作成

-- ===========================================
-- ENUM 型
-- ===========================================

CREATE TYPE oauth_provider AS ENUM ('google');
CREATE TYPE gender AS ENUM ('male', 'female', 'neutral');
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- ===========================================
-- メディア関連テーブル
-- ===========================================

-- 画像
CREATE TABLE images (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	mime_type VARCHAR(100) NOT NULL,
	path VARCHAR(1024) NOT NULL,
	filename VARCHAR(255) NOT NULL,
	file_size INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN images.path IS 'GCS 上のパス（例: images/xxx.png）';

-- 音声
CREATE TABLE audios (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	mime_type VARCHAR(100) NOT NULL,
	path VARCHAR(1024) NOT NULL,
	filename VARCHAR(255) NOT NULL,
	file_size INTEGER NOT NULL,
	duration_ms INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- マスターテーブル
-- ===========================================

-- ボイス
CREATE TABLE voices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	provider VARCHAR(50) NOT NULL,
	provider_voice_id VARCHAR(100) NOT NULL,
	name VARCHAR(100) NOT NULL,
	gender gender,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (provider, provider_voice_id)
);

CREATE INDEX idx_voices_provider ON voices (provider);
CREATE INDEX idx_voices_is_active ON voices (is_active);

-- カテゴリ
CREATE TABLE categories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug VARCHAR(50) NOT NULL UNIQUE,
	name VARCHAR(100) NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_sort_order ON categories (sort_order);
CREATE INDEX idx_categories_is_active ON categories (is_active);

-- システム BGM（マスタ）
CREATE TABLE system_bgms (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	audio_id UUID NOT NULL REFERENCES audios (id) ON DELETE RESTRICT,
	name VARCHAR(255) NOT NULL UNIQUE,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_bgms_sort_order ON system_bgms (sort_order);
CREATE INDEX idx_system_bgms_is_active ON system_bgms (is_active);
CREATE INDEX idx_system_bgms_audio_id ON system_bgms (audio_id);

-- ===========================================
-- 認証関連テーブル
-- ===========================================

-- ユーザー
CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	email VARCHAR(255) NOT NULL UNIQUE,
	username VARCHAR(20) NOT NULL UNIQUE,
	display_name VARCHAR(20) NOT NULL,
	avatar_id UUID REFERENCES images (id) ON DELETE SET NULL,
	role user_role NOT NULL DEFAULT 'user',
	user_prompt TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- パスワード認証情報
CREATE TABLE credentials (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
	password_hash VARCHAR(255) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- OAuth アカウント
CREATE TABLE oauth_accounts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	provider oauth_provider NOT NULL,
	provider_user_id VARCHAR(255) NOT NULL,
	access_token VARCHAR(1024),
	refresh_token VARCHAR(1024),
	expires_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts (user_id);

-- ===========================================
-- ユーザーデータテーブル
-- ===========================================

-- BGM（ユーザー所有）
-- channels テーブルが参照するため、channels より先に作成
CREATE TABLE bgms (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	audio_id UUID NOT NULL REFERENCES audios (id) ON DELETE RESTRICT,
	name VARCHAR(255) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, name)
);

CREATE INDEX idx_bgms_user_id ON bgms (user_id);
CREATE INDEX idx_bgms_audio_id ON bgms (audio_id);

-- チャンネル
CREATE TABLE channels (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	description TEXT NOT NULL,
	artwork_id UUID REFERENCES images (id) ON DELETE SET NULL,
	category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
	default_bgm_id UUID REFERENCES bgms (id) ON DELETE SET NULL,
	default_system_bgm_id UUID REFERENCES system_bgms (id) ON DELETE SET NULL,
	published_at TIMESTAMP,
	user_prompt TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- default_bgm_id と default_system_bgm_id は同時に設定不可
	CONSTRAINT chk_channels_default_bgm_exclusive CHECK (NOT (default_bgm_id IS NOT NULL AND default_system_bgm_id IS NOT NULL))
);

CREATE INDEX idx_channels_user_id ON channels (user_id);
CREATE INDEX idx_channels_category_id ON channels (category_id);
CREATE INDEX idx_channels_published_at ON channels (published_at);
CREATE INDEX idx_channels_default_bgm_id ON channels (default_bgm_id);
CREATE INDEX idx_channels_default_system_bgm_id ON channels (default_system_bgm_id);

-- キャラクター（ユーザー所有）
CREATE TABLE characters (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	name VARCHAR(255) NOT NULL,
	persona TEXT NOT NULL,
	voice_id UUID NOT NULL REFERENCES voices (id) ON DELETE RESTRICT,
	avatar_id UUID REFERENCES images (id) ON DELETE SET NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, name)
);

CREATE INDEX idx_characters_user_id ON characters (user_id);
CREATE INDEX idx_characters_avatar_id ON characters (avatar_id);

-- チャンネルとキャラクターの中間テーブル
CREATE TABLE channel_characters (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	channel_id UUID NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
	character_id UUID NOT NULL REFERENCES characters (id) ON DELETE RESTRICT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (channel_id, character_id)
);

CREATE INDEX idx_channel_characters_channel_id ON channel_characters (channel_id);
CREATE INDEX idx_channel_characters_character_id ON channel_characters (character_id);

-- エピソード
CREATE TABLE episodes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	channel_id UUID NOT NULL REFERENCES channels (id) ON DELETE CASCADE,
	title VARCHAR(255) NOT NULL,
	description TEXT NOT NULL,
	bgm_id UUID REFERENCES bgms (id) ON DELETE SET NULL,
	system_bgm_id UUID REFERENCES system_bgms (id) ON DELETE SET NULL,
	full_audio_id UUID REFERENCES audios (id) ON DELETE SET NULL,
	published_at TIMESTAMP,
	user_prompt TEXT NOT NULL DEFAULT '',
	voice_style TEXT NOT NULL DEFAULT '',
	artwork_id UUID REFERENCES images (id) ON DELETE SET NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- bgm_id と system_bgm_id は同時に設定不可
	CONSTRAINT chk_episodes_bgm_exclusive CHECK (NOT (bgm_id IS NOT NULL AND system_bgm_id IS NOT NULL))
);

CREATE INDEX idx_episodes_channel_id ON episodes (channel_id);
CREATE INDEX idx_episodes_published_at ON episodes (published_at);

-- 台本行
CREATE TABLE script_lines (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	line_order INTEGER NOT NULL,
	speaker_id UUID NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
	text TEXT NOT NULL,
	emotion TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (episode_id, line_order) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_script_lines_episode_id ON script_lines (episode_id);

-- ===========================================
-- ユーザーインタラクションテーブル
-- ===========================================

-- いいね
CREATE TABLE likes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_likes_user_id ON likes (user_id);
CREATE INDEX idx_likes_episode_id ON likes (episode_id);

-- ブックマーク
CREATE TABLE bookmarks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_bookmarks_user_id ON bookmarks (user_id);
CREATE INDEX idx_bookmarks_episode_id ON bookmarks (episode_id);

-- 再生履歴
CREATE TABLE playback_histories (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	progress_ms INTEGER NOT NULL DEFAULT 0,
	completed BOOLEAN NOT NULL DEFAULT false,
	played_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_playback_histories_user_id ON playback_histories (user_id);
CREATE INDEX idx_playback_histories_episode_id ON playback_histories (episode_id);
CREATE INDEX idx_playback_histories_user_id_played_at ON playback_histories (user_id, played_at);

-- フォロー
CREATE TABLE follows (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_follows_user_id ON follows (user_id);
CREATE INDEX idx_follows_episode_id ON follows (episode_id);
