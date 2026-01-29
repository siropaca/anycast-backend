-- 全テーブルのデータを削除する
-- seed 投入前にクリーンな状態にするため、全テーブルを TRUNCATE する
-- マスターテーブル（voices, categories）と system_bgms は migration で投入されるため除外

TRUNCATE
	comments,
	follows,
	playback_histories,
	playlist_items,
	playlists,
	reactions,
	script_jobs,
	audio_jobs,
	script_lines,
	episodes,
	channel_characters,
	channels,
	bgms,
	characters,
	oauth_accounts,
	credentials,
	users,
	images,
	audios
CASCADE;
