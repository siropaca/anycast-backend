-- テストユーザーを作成する（開発環境専用）
-- パスワード: password

-- 既存のテストユーザーを削除（冪等性のため）
DELETE FROM credentials WHERE user_id IN (
    SELECT id FROM users WHERE email IN ('test@example.com', 'alice@example.com')
);
DELETE FROM users WHERE email IN ('test@example.com', 'alice@example.com');

-- テストユーザーを作成
INSERT INTO users (id, email, name) VALUES
    ('8def69af-dae9-4641-a0e5-100107626933', 'test@example.com', 'Test User'),
    ('63974c18-b3ca-4e38-bb34-325f35cb3891', 'alice@example.com', 'Alice');

-- パスワード認証情報を作成（パスワード: password）
INSERT INTO credentials (user_id, password_hash) VALUES
    ('8def69af-dae9-4641-a0e5-100107626933', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
    ('63974c18-b3ca-4e38-bb34-325f35cb3891', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO');
