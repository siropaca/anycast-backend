-- システム BGM の seed データ

-- 音声レコードを挿入
INSERT INTO audios (mime_type, path, filename, file_size, duration_ms) VALUES
	('audio/mpeg', 'system/You_and_Me.mp3', 'You_and_Me.mp3', 4508877, 108000),
	('audio/mpeg', 'system/2_23_AM.mp3', '2_23_AM.mp3', 8074035, 192000),
	('audio/mpeg', 'system/pastel_house.mp3', 'パステルハウス.mp3', 7549747, 180000),
	('audio/mpeg', 'system/野良猫は宇宙を目指した.mp3', '野良猫は宇宙を目指した.mp3', 1992295, 48000);

-- システム BGM レコードを挿入（audio_id はパスで参照）
INSERT INTO system_bgms (audio_id, name, sort_order) VALUES
	((SELECT id FROM audios WHERE path = 'system/You_and_Me.mp3'), 'You and Me', 1),
	((SELECT id FROM audios WHERE path = 'system/2_23_AM.mp3'), '2:23 AM', 2),
	((SELECT id FROM audios WHERE path = 'system/pastel_house.mp3'), 'パステルハウス', 3),
	((SELECT id FROM audios WHERE path = 'system/野良猫は宇宙を目指した.mp3'), '野良猫は宇宙を目指した', 4);
