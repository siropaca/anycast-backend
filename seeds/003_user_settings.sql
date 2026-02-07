-- ユーザー設定を適用する

-- 全ユーザーにアバターを設定
UPDATE users SET avatar_id = '014e00d8-fc7a-4a64-941d-5cb4edaa901a' WHERE username = 'test_user';
UPDATE users SET avatar_id = '507fa3f3-424f-41ab-8c73-bdb302a9d82f' WHERE username = 'test_user2';
UPDATE users SET avatar_id = '1096535e-1c18-4761-9325-91675e30a989' WHERE username = 'test_user3';
UPDATE users SET avatar_id = '6621568b-4d72-438e-8eb8-6d90752b6120' WHERE username = 'test_user4';
UPDATE users SET avatar_id = 'c799427a-9a6c-4b40-9e3e-55e72b2a6c8b' WHERE username = 'test_user5';
UPDATE users SET avatar_id = 'eef7ab18-a807-494e-ba70-f099022ebb71' WHERE username = 'test_user6';
UPDATE users SET avatar_id = '7fe1f005-04d0-4ea8-b9f2-9140129c47c4' WHERE username = 'test_user7';
UPDATE users SET avatar_id = '5b0cb42d-18b6-4f9f-a48e-66650c1d3b5e' WHERE username = 'test_user8';
UPDATE users SET avatar_id = '76084039-1adc-4487-a55f-24de2657a222' WHERE username = 'test_user9';
UPDATE users SET avatar_id = 'e8ff3f67-40c0-4323-94cf-053fc3dd5376' WHERE username = 'test_user10';

-- 全ユーザーに自己紹介文を設定
UPDATE users SET bio = 'テスト用アカウントです。よろしくお願いします！' WHERE username = 'test_user';
UPDATE users SET bio = 'ポッドキャスト初心者です。いろいろ試しています。' WHERE username = 'test_user2';
UPDATE users SET bio = 'テクノロジーと科学が好きです。' WHERE username = 'test_user3';
UPDATE users SET bio = 'ビジネス系の話題を中心に配信しています。' WHERE username = 'test_user4';
UPDATE users SET bio = '日常のゆるいトークをお届けします。' WHERE username = 'test_user5';
UPDATE users SET bio = 'ニュースや時事問題について語ります。' WHERE username = 'test_user6';
UPDATE users SET bio = '映画とアートが大好きです。' WHERE username = 'test_user7';
UPDATE users SET bio = 'スポーツ観戦が趣味です。' WHERE username = 'test_user8';
UPDATE users SET bio = '健康と音楽についてお話しします。' WHERE username = 'test_user9';
UPDATE users SET bio = '歴史と教育に興味があります。' WHERE username = 'test_user10';

-- 全ユーザーにヘッダー画像を設定
UPDATE users SET header_image_id = 'a1b2c3d4-1111-4000-8000-000000000001' WHERE username = 'test_user';
UPDATE users SET header_image_id = 'a1b2c3d4-2222-4000-8000-000000000002' WHERE username = 'test_user2';
UPDATE users SET header_image_id = 'a1b2c3d4-3333-4000-8000-000000000003' WHERE username = 'test_user3';
UPDATE users SET header_image_id = 'a1b2c3d4-4444-4000-8000-000000000004' WHERE username = 'test_user4';
UPDATE users SET header_image_id = 'a1b2c3d4-5555-4000-8000-000000000005' WHERE username = 'test_user5';
UPDATE users SET header_image_id = 'a1b2c3d4-6666-4000-8000-000000000006' WHERE username = 'test_user6';
UPDATE users SET header_image_id = 'a1b2c3d4-7777-4000-8000-000000000007' WHERE username = 'test_user7';
UPDATE users SET header_image_id = 'a1b2c3d4-8888-4000-8000-000000000008' WHERE username = 'test_user8';
UPDATE users SET header_image_id = 'a1b2c3d4-9999-4000-8000-000000000009' WHERE username = 'test_user9';
UPDATE users SET header_image_id = 'a1b2c3d4-aaaa-4000-8000-00000000000a' WHERE username = 'test_user10';
