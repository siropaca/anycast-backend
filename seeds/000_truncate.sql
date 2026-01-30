-- 全テーブルのデータを削除する
-- seed 投入前にクリーンな状態にするため、全テーブルを TRUNCATE する
-- マスターテーブル（voices, categories）と system_bgms は migration で投入されるため除外
-- audios は system_bgms が FK 参照しているため TRUNCATE CASCADE から除外し、
-- system_bgms に紐づかないレコードのみ DELETE する

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
	images
CASCADE;

-- system_bgms が参照する audios は残し、それ以外を削除
DELETE FROM audios WHERE id NOT IN (SELECT audio_id FROM system_bgms);
