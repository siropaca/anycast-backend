-- Anycast データベーススキーマ
-- 全テーブルを最終形の状態で作成

-- ===========================================
-- ENUM 型
-- ===========================================

CREATE TYPE oauth_provider AS ENUM ('google');
CREATE TYPE gender AS ENUM ('male', 'female', 'neutral');
CREATE TYPE user_role AS ENUM ('user', 'admin');
CREATE TYPE audio_job_status AS ENUM ('pending', 'processing', 'canceling', 'completed', 'failed', 'canceled');
CREATE TYPE audio_job_type AS ENUM ('voice', 'full', 'remix');
CREATE TYPE script_job_status AS ENUM ('pending', 'processing', 'canceling', 'completed', 'failed', 'canceled');
CREATE TYPE reaction_type AS ENUM ('like', 'bad');
CREATE TYPE contact_category AS ENUM ('general', 'bug_report', 'feature_request', 'other');

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
	sample_audio_url VARCHAR(1024) NOT NULL,
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
	image_id UUID REFERENCES images (id) ON DELETE SET NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_categories_sort_order ON categories (sort_order);
CREATE INDEX idx_categories_is_active ON categories (is_active);
CREATE INDEX idx_categories_image_id ON categories (image_id);

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
	header_image_id UUID REFERENCES images (id) ON DELETE SET NULL,
	bio VARCHAR(200) NOT NULL DEFAULT '',
	role user_role NOT NULL DEFAULT 'user',
	user_prompt VARCHAR(2000) NOT NULL DEFAULT '',
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

-- リフレッシュトークン
CREATE TABLE refresh_tokens (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	token VARCHAR(255) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

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
	description VARCHAR(2000) NOT NULL,
	artwork_id UUID REFERENCES images (id) ON DELETE SET NULL,
	category_id UUID NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
	default_bgm_id UUID REFERENCES bgms (id) ON DELETE SET NULL,
	default_system_bgm_id UUID REFERENCES system_bgms (id) ON DELETE SET NULL,
	published_at TIMESTAMP,
	user_prompt VARCHAR(2000),
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
	persona VARCHAR(2000) NOT NULL,
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
	description VARCHAR(2000) NOT NULL,
	bgm_id UUID REFERENCES bgms (id) ON DELETE SET NULL,
	system_bgm_id UUID REFERENCES system_bgms (id) ON DELETE SET NULL,
	voice_audio_id UUID REFERENCES audios (id) ON DELETE SET NULL,
	full_audio_id UUID REFERENCES audios (id) ON DELETE SET NULL,
	audio_outdated BOOLEAN NOT NULL DEFAULT false,
	play_count INTEGER NOT NULL DEFAULT 0,
	published_at TIMESTAMP,
	voice_style TEXT NOT NULL DEFAULT '',
	artwork_id UUID REFERENCES images (id) ON DELETE SET NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	-- bgm_id と system_bgm_id は同時に設定不可
	CONSTRAINT chk_episodes_bgm_exclusive CHECK (NOT (bgm_id IS NOT NULL AND system_bgm_id IS NOT NULL))
);

CREATE INDEX idx_episodes_channel_id ON episodes (channel_id);
CREATE INDEX idx_episodes_published_at ON episodes (published_at);

-- 音声生成ジョブ
CREATE TABLE audio_jobs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	status audio_job_status NOT NULL DEFAULT 'pending',
	job_type audio_job_type NOT NULL DEFAULT 'voice',
	progress INTEGER NOT NULL DEFAULT 0,
	voice_style TEXT NOT NULL DEFAULT '',
	-- BGM 参照
	bgm_id UUID REFERENCES bgms (id) ON DELETE SET NULL,
	system_bgm_id UUID REFERENCES system_bgms (id) ON DELETE SET NULL,
	-- BGM ミキシング設定
	bgm_volume_db DECIMAL(5, 2) NOT NULL DEFAULT -15.0,
	fade_out_ms INTEGER NOT NULL DEFAULT 3000,
	padding_start_ms INTEGER NOT NULL DEFAULT 500,
	padding_end_ms INTEGER NOT NULL DEFAULT 1000,
	-- 結果
	result_audio_id UUID REFERENCES audios (id) ON DELETE SET NULL,
	error_message TEXT,
	error_code VARCHAR(50),
	-- タイムスタンプ
	started_at TIMESTAMP,
	completed_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT chk_audio_jobs_bgm_exclusive CHECK (NOT (bgm_id IS NOT NULL AND system_bgm_id IS NOT NULL))
);

CREATE INDEX idx_audio_jobs_episode_id ON audio_jobs (episode_id);
CREATE INDEX idx_audio_jobs_user_id ON audio_jobs (user_id);
CREATE INDEX idx_audio_jobs_status ON audio_jobs (status);
CREATE INDEX idx_audio_jobs_created_at ON audio_jobs (created_at DESC);

-- 台本生成ジョブ
CREATE TABLE script_jobs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	status script_job_status NOT NULL DEFAULT 'pending',
	progress INTEGER NOT NULL DEFAULT 0,
	-- 生成パラメータ
	prompt VARCHAR(2000) NOT NULL,
	duration_minutes INTEGER NOT NULL DEFAULT 10,
	with_emotion BOOLEAN NOT NULL DEFAULT false,
	-- 結果
	error_message TEXT,
	error_code VARCHAR(50),
	-- タイムスタンプ
	started_at TIMESTAMP,
	completed_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_script_jobs_episode_id ON script_jobs (episode_id);
CREATE INDEX idx_script_jobs_user_id ON script_jobs (user_id);
CREATE INDEX idx_script_jobs_status ON script_jobs (status);
CREATE INDEX idx_script_jobs_created_at ON script_jobs (created_at DESC);

-- 台本行
CREATE TABLE script_lines (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	line_order INTEGER NOT NULL,
	speaker_id UUID NOT NULL REFERENCES characters (id) ON DELETE CASCADE,
	text VARCHAR(500) NOT NULL,
	emotion VARCHAR(20),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (episode_id, line_order) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_script_lines_episode_id ON script_lines (episode_id);

-- ===========================================
-- ユーザーインタラクションテーブル
-- ===========================================

-- リアクション
CREATE TABLE reactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	reaction_type reaction_type NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_reactions_user_id ON reactions (user_id);
CREATE INDEX idx_reactions_episode_id ON reactions (episode_id);

-- 再生リスト
CREATE TABLE playlists (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	name VARCHAR(100) NOT NULL,
	description VARCHAR(500) NOT NULL DEFAULT '',
	is_default BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, name)
);

CREATE INDEX idx_playlists_user_id ON playlists (user_id);
CREATE UNIQUE INDEX idx_playlists_user_id_default ON playlists (user_id) WHERE is_default = true;

-- 再生リストアイテム
CREATE TABLE playlist_items (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	playlist_id UUID NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (playlist_id, episode_id),
	UNIQUE (playlist_id, position) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_playlist_items_playlist_id ON playlist_items (playlist_id);
CREATE INDEX idx_playlist_items_episode_id ON playlist_items (episode_id);

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
	target_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, target_user_id),
	CONSTRAINT chk_follows_no_self_follow CHECK (user_id != target_user_id)
);

CREATE INDEX idx_follows_user_id ON follows (user_id);
CREATE INDEX idx_follows_target_user_id ON follows (target_user_id);

-- コメント
CREATE TABLE comments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	episode_id UUID NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	deleted_at TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT chk_comments_content_length CHECK (char_length(content) >= 1 AND char_length(content) <= 1000)
);

CREATE INDEX idx_comments_user_id ON comments (user_id);
CREATE INDEX idx_comments_episode_id ON comments (episode_id);
CREATE INDEX idx_comments_created_at ON comments (created_at DESC);
CREATE INDEX idx_comments_deleted_at ON comments (deleted_at);

-- ボイスお気に入り
CREATE TABLE favorite_voices (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	voice_id UUID NOT NULL REFERENCES voices (id) ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE (user_id, voice_id)
);

CREATE INDEX idx_favorite_voices_user_id ON favorite_voices (user_id);
CREATE INDEX idx_favorite_voices_voice_id ON favorite_voices (voice_id);

-- フィードバック
CREATE TABLE feedbacks (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	screenshot_id UUID REFERENCES images (id) ON DELETE SET NULL,
	page_url VARCHAR(2048),
	user_agent VARCHAR(1024),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT chk_feedbacks_content_length CHECK (char_length(content) >= 1 AND char_length(content) <= 5000)
);

CREATE INDEX idx_feedbacks_user_id ON feedbacks (user_id);
CREATE INDEX idx_feedbacks_created_at ON feedbacks (created_at DESC);

-- お問い合わせ
CREATE TABLE contacts (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	user_id UUID REFERENCES users (id) ON DELETE SET NULL,
	category contact_category NOT NULL,
	email VARCHAR(255) NOT NULL,
	name VARCHAR(100) NOT NULL,
	content VARCHAR(5000) NOT NULL,
	user_agent VARCHAR(1024),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT chk_contacts_content_length CHECK (char_length(content) >= 1 AND char_length(content) <= 5000)
);

CREATE INDEX idx_contacts_user_id ON contacts (user_id);
CREATE INDEX idx_contacts_category ON contacts (category);
CREATE INDEX idx_contacts_created_at ON contacts (created_at DESC);
