-- テストユーザーを作成する
-- パスワード: password

-- 既存のテストユーザーを削除（冪等性のため）
-- script_lines.speaker_id は ON DELETE RESTRICT のため、先に削除する
DELETE FROM script_lines WHERE episode_id IN (
	SELECT e.id FROM episodes e
	JOIN channels c ON e.channel_id = c.id
	JOIN users u ON c.user_id = u.id
	WHERE u.username IN ('test_user', 'test_user2')
);
-- credentials, channels 等は ON DELETE CASCADE で自動削除される
DELETE FROM users WHERE username IN ('test_user', 'test_user2');

-- テストユーザーを作成
INSERT INTO users (id, email, username, display_name) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', 'test@example.com', 'test_user', 'Test User'),
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', 'test2@example.com', 'test_user2', 'Test User 2');

-- パスワード認証情報を作成（パスワード: password）
INSERT INTO credentials (user_id, password_hash) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO');
