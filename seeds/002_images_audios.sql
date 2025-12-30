-- テスト用の画像・音声データを作成する

-- 既存のテストデータを削除（冪等性のため）
DELETE FROM images WHERE url LIKE 'https://storage.googleapis.com/anycast-dev/%';
DELETE FROM audios WHERE url LIKE 'https://storage.googleapis.com/anycast-dev/%';

-- テスト用画像を作成
INSERT INTO images (id, mime_type, url, filename, file_size) VALUES
	('4946f33c-3c66-40ca-8b35-3bbdfe65b20c', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-artwork1.png', 'test-artwork1.png', 72090),
	('9ee172c8-6deb-4598-a379-d7fdf502db9a', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-artwork2.png', 'test-artwork2.png', 72090),
	('cb12d606-f704-49b4-8797-254dfada4123', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-icon.png', 'test-icon.png', 337306);

-- テスト用音声を作成
INSERT INTO audios (id, mime_type, url, filename, file_size, duration_ms) VALUES
	('4b33e6ee-c81b-4795-a843-74fd82fa4fd2', 'audio/mpeg', 'https://storage.googleapis.com/anycast-dev/audios/test-audio1.mp3', 'test-audio1.mp3', 47514, 4000);
