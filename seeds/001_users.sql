-- テストユーザーを作成する
-- パスワード: password

-- テストユーザーを作成
INSERT INTO users (id, email, username, display_name) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', 'test@example.com', 'test_user', 'Test User'),
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', 'test2@example.com', 'test_user2', 'Test User 2'),
	('4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', 'test3@example.com', 'test_user3', 'Sakura'),
	('d6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', 'test4@example.com', 'test_user4', 'Ren'),
	('b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', 'test5@example.com', 'test_user5', 'Hina'),
	('80adf759-b01c-4726-87b4-7b9c659483a4', 'test6@example.com', 'test_user6', 'Kaito'),
	('8450d256-8630-4044-8a69-fc8671e6e5c1', 'test7@example.com', 'test_user7', 'Mio'),
	('7d4e20d4-98ca-4b79-901d-c51e43c38e2f', 'test8@example.com', 'test_user8', 'Yuto'),
	('c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', 'test9@example.com', 'test_user9', 'Aoi'),
	('767b5ed0-a663-437a-9cc9-b8cef6d0731e', 'test10@example.com', 'test_user10', 'Sora');

-- パスワード認証情報を作成（パスワード: password）
INSERT INTO credentials (user_id, password_hash) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('d6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('80adf759-b01c-4726-87b4-7b9c659483a4', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('8450d256-8630-4044-8a69-fc8671e6e5c1', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('7d4e20d4-98ca-4b79-901d-c51e43c38e2f', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO'),
	('767b5ed0-a663-437a-9cc9-b8cef6d0731e', '$2a$10$MnwjdQyqBX4NlH2birda6uSeGDlE/q4Vjv8fugcnfCOZE/ERlORxO');
