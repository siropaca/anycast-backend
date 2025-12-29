-- テストユーザーを作成する（開発環境専用）
-- パスワード: password

-- 既存のテストユーザーを削除（冪等性のため）
-- credentials は ON DELETE CASCADE で自動削除される
DELETE FROM users WHERE username IN ('test_user');

-- テストユーザーを作成
INSERT INTO users (id, email, username, display_name) VALUES
    ('8def69af-dae9-4641-a0e5-100107626933', 'test@example.com', 'test_user', 'Test User');

-- パスワード認証情報を作成（パスワード: password）
INSERT INTO credentials (user_id, password_hash) VALUES
    ('8def69af-dae9-4641-a0e5-100107626933', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO');
