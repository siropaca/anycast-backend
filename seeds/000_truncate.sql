-- 全テーブルのデータを削除する
-- seed 投入前にクリーンな状態にするため、全テーブルを TRUNCATE する
-- マスターテーブル（voices, categories）と system_bgms は migration で投入されるため除外
-- images は categories が FK 参照しているため TRUNCATE CASCADE から除外し、
-- マスターテーブルに紐づかないレコードのみ DELETE する
-- audios は system_bgms が FK 参照しているため同様に除外

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
	users
CASCADE;

-- categories が参照する images は残し、それ以外を削除
DELETE FROM images WHERE id NOT IN (
	SELECT image_id FROM categories WHERE image_id IS NOT NULL
);

-- system_bgms が参照する audios は残し、それ以外を削除
DELETE FROM audios WHERE id NOT IN (SELECT audio_id FROM system_bgms);
