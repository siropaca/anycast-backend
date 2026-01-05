-- テスト用の画像データを作成する

-- 既存のテストデータを削除（冪等性のため）
DELETE FROM images WHERE url LIKE 'https://storage.googleapis.com/anycast-dev/%';

-- テスト用画像を作成
INSERT INTO images (id, mime_type, url, filename, file_size) VALUES
	('4946f33c-3c66-40ca-8b35-3bbdfe65b20c', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-artwork1.png', 'test-artwork1.png', 72090),
	('9ee172c8-6deb-4598-a379-d7fdf502db9a', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-artwork2.png', 'test-artwork2.png', 72090),
	('cb12d606-f704-49b4-8797-254dfada4123', 'image/png', 'https://storage.googleapis.com/anycast-dev/images/test-icon.png', 'test-icon.png', 337306);
