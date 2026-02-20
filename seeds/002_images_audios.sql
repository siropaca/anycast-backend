-- テスト用の画像データを作成する
-- picsum.photos/seed を利用した外部 URL 画像

-- ===========================================
-- ユーザーアバター画像（10枚）
-- ===========================================
INSERT INTO images (id, mime_type, path, filename, file_size) VALUES
	('014e00d8-fc7a-4a64-941d-5cb4edaa901a', 'image/jpeg', 'https://picsum.photos/seed/avatar-user1/200/200', 'avatar-user1.jpg', 0),
	('507fa3f3-424f-41ab-8c73-bdb302a9d82f', 'image/jpeg', 'https://picsum.photos/seed/avatar-user2/200/200', 'avatar-user2.jpg', 0),
	('1096535e-1c18-4761-9325-91675e30a989', 'image/jpeg', 'https://picsum.photos/seed/avatar-user3/200/200', 'avatar-user3.jpg', 0),
	('6621568b-4d72-438e-8eb8-6d90752b6120', 'image/jpeg', 'https://picsum.photos/seed/avatar-user4/200/200', 'avatar-user4.jpg', 0),
	('c799427a-9a6c-4b40-9e3e-55e72b2a6c8b', 'image/jpeg', 'https://picsum.photos/seed/avatar-user5/200/200', 'avatar-user5.jpg', 0),
	('eef7ab18-a807-494e-ba70-f099022ebb71', 'image/jpeg', 'https://picsum.photos/seed/avatar-user6/200/200', 'avatar-user6.jpg', 0),
	('7fe1f005-04d0-4ea8-b9f2-9140129c47c4', 'image/jpeg', 'https://picsum.photos/seed/avatar-user7/200/200', 'avatar-user7.jpg', 0),
	('5b0cb42d-18b6-4f9f-a48e-66650c1d3b5e', 'image/jpeg', 'https://picsum.photos/seed/avatar-user8/200/200', 'avatar-user8.jpg', 0),
	('76084039-1adc-4487-a55f-24de2657a222', 'image/jpeg', 'https://picsum.photos/seed/avatar-user9/200/200', 'avatar-user9.jpg', 0),
	('e8ff3f67-40c0-4323-94cf-053fc3dd5376', 'image/jpeg', 'https://picsum.photos/seed/avatar-user10/200/200', 'avatar-user10.jpg', 0);

-- ===========================================
-- ユーザーヘッダー画像（10枚）
-- ===========================================
INSERT INTO images (id, mime_type, path, filename, file_size) VALUES
	('a1b2c3d4-1111-4000-8000-000000000001', 'image/jpeg', 'https://picsum.photos/seed/header-user1/1500/500', 'header-user1.jpg', 0),
	('a1b2c3d4-2222-4000-8000-000000000002', 'image/jpeg', 'https://picsum.photos/seed/header-user2/1500/500', 'header-user2.jpg', 0),
	('a1b2c3d4-3333-4000-8000-000000000003', 'image/jpeg', 'https://picsum.photos/seed/header-user3/1500/500', 'header-user3.jpg', 0),
	('a1b2c3d4-4444-4000-8000-000000000004', 'image/jpeg', 'https://picsum.photos/seed/header-user4/1500/500', 'header-user4.jpg', 0),
	('a1b2c3d4-5555-4000-8000-000000000005', 'image/jpeg', 'https://picsum.photos/seed/header-user5/1500/500', 'header-user5.jpg', 0),
	('a1b2c3d4-6666-4000-8000-000000000006', 'image/jpeg', 'https://picsum.photos/seed/header-user6/1500/500', 'header-user6.jpg', 0),
	('a1b2c3d4-7777-4000-8000-000000000007', 'image/jpeg', 'https://picsum.photos/seed/header-user7/1500/500', 'header-user7.jpg', 0),
	('a1b2c3d4-8888-4000-8000-000000000008', 'image/jpeg', 'https://picsum.photos/seed/header-user8/1500/500', 'header-user8.jpg', 0),
	('a1b2c3d4-9999-4000-8000-000000000009', 'image/jpeg', 'https://picsum.photos/seed/header-user9/1500/500', 'header-user9.jpg', 0),
	('a1b2c3d4-aaaa-4000-8000-00000000000a', 'image/jpeg', 'https://picsum.photos/seed/header-user10/1500/500', 'header-user10.jpg', 0);

-- ===========================================
-- チャンネルアートワーク画像（15枚）
-- ===========================================
INSERT INTO images (id, mime_type, path, filename, file_size) VALUES
	('6852eda5-db2e-42f0-9870-c00a396e6bf8', 'image/jpeg', 'https://picsum.photos/seed/ch-techtalk/400/400', 'ch-techtalk.jpg', 0),
	('a28d9fb9-1c8f-4835-a0ad-96afdb6e3279', 'image/jpeg', 'https://picsum.photos/seed/ch-yurufuwa/400/400', 'ch-yurufuwa.jpg', 0),
	('ef89720a-9e2e-4021-8c28-f4e42d1c3d93', 'image/jpeg', 'https://picsum.photos/seed/ch-business/400/400', 'ch-business.jpg', 0),
	('3cbba00a-e754-4516-bf68-958c84c288e3', 'image/jpeg', 'https://picsum.photos/seed/ch-science/400/400', 'ch-science.jpg', 0),
	('9f17a722-a320-4f6f-a827-05acf96c36c9', 'image/jpeg', 'https://picsum.photos/seed/ch-honobono/400/400', 'ch-honobono.jpg', 0),
	('3084384d-2e85-437a-adfd-a15a0b51414d', 'image/jpeg', 'https://picsum.photos/seed/ch-news/400/400', 'ch-news.jpg', 0),
	('d259698b-7f50-488b-b2c6-5d822be5645d', 'image/jpeg', 'https://picsum.photos/seed/ch-movie/400/400', 'ch-movie.jpg', 0),
	('e9c8d1d7-f953-4f3f-a960-6df46ab6eaff', 'image/jpeg', 'https://picsum.photos/seed/ch-art/400/400', 'ch-art.jpg', 0),
	('f562344f-1bf0-4767-8dd3-34743c92c101', 'image/jpeg', 'https://picsum.photos/seed/ch-sports/400/400', 'ch-sports.jpg', 0),
	('eb07435e-b994-422f-9d35-f8c7b9e437d3', 'image/jpeg', 'https://picsum.photos/seed/ch-healthy/400/400', 'ch-healthy.jpg', 0),
	('90d6d166-3dc0-41f1-bf5e-78c11b25598c', 'image/jpeg', 'https://picsum.photos/seed/ch-music/400/400', 'ch-music.jpg', 0),
	('1731f558-78bd-4caa-bdc1-aaf7f4875f87', 'image/jpeg', 'https://picsum.photos/seed/ch-history/400/400', 'ch-history.jpg', 0),
	('bb6b9e97-2a5e-4bf0-bc60-814d8cbd99d9', 'image/jpeg', 'https://picsum.photos/seed/ch-comedy/400/400', 'ch-comedy.jpg', 0),
	('ec3b5427-e883-466b-a98d-0f3bc0084da1', 'image/jpeg', 'https://picsum.photos/seed/ch-education/400/400', 'ch-education.jpg', 0),
	('bdf2b1f0-735a-489b-be3b-97c97c06a2c8', 'image/jpeg', 'https://picsum.photos/seed/ch-fiction/400/400', 'ch-fiction.jpg', 0),
	('50100001-0005-4000-a000-000000000001', 'image/jpeg', 'https://picsum.photos/seed/ch-solo-talk/400/400', 'ch-solo-talk.jpg', 0);

-- ===========================================
-- エピソードアートワーク画像（36枚）
-- ===========================================
INSERT INTO images (id, mime_type, path, filename, file_size) VALUES
	('209e8d96-dbc1-4dcf-a747-a0ac4ec37a0c', 'image/jpeg', 'https://picsum.photos/seed/ep-ai-future/400/400', 'ep-ai-future.jpg', 0),
	('e463cd78-0e09-4d6b-ac83-ec01fb25d5a9', 'image/jpeg', 'https://picsum.photos/seed/ep-smarthome/400/400', 'ep-smarthome.jpg', 0),
	('783ac817-5deb-4711-a8d1-b3e9721b891c', 'image/jpeg', 'https://picsum.photos/seed/ep-hobby/400/400', 'ep-hobby.jpg', 0),
	('45886285-2b5b-4575-8c5c-9c95d139a803', 'image/jpeg', 'https://picsum.photos/seed/ep-startup/400/400', 'ep-startup.jpg', 0),
	('2a2dacc8-670e-4d3d-965c-9c8ae6679bd5', 'image/jpeg', 'https://picsum.photos/seed/ep-funding/400/400', 'ep-funding.jpg', 0),
	('aae14562-a3a1-4507-8f24-bb72a11dda5c', 'image/jpeg', 'https://picsum.photos/seed/ep-quantum/400/400', 'ep-quantum.jpg', 0),
	('9a7dc980-2da8-4bf1-9557-9a52193af39b', 'image/jpeg', 'https://picsum.photos/seed/ep-space/400/400', 'ep-space.jpg', 0),
	('ccaedc4a-59bc-4736-980e-699d85dd822b', 'image/jpeg', 'https://picsum.photos/seed/ep-dna/400/400', 'ep-dna.jpg', 0),
	('21188860-1bc2-45f2-a15c-10462e2c20cc', 'image/jpeg', 'https://picsum.photos/seed/ep-cafe/400/400', 'ep-cafe.jpg', 0),
	('8b53c1a7-a536-4acb-8a61-aaf6eea880db', 'image/jpeg', 'https://picsum.photos/seed/ep-holiday/400/400', 'ep-holiday.jpg', 0),
	('679e7f0d-ca8a-46ca-b7c5-a4c44ddecc28', 'image/jpeg', 'https://picsum.photos/seed/ep-sns/400/400', 'ep-sns.jpg', 0),
	('467ae4ef-5302-4750-b71e-12a74d5c9860', 'image/jpeg', 'https://picsum.photos/seed/ep-election/400/400', 'ep-election.jpg', 0),
	('1971d88a-91f5-4b72-9e0b-9319fe0c3e1a', 'image/jpeg', 'https://picsum.photos/seed/ep-climate/400/400', 'ep-climate.jpg', 0),
	('4010aa53-c71b-4640-94c2-283cf3ff0a02', 'image/jpeg', 'https://picsum.photos/seed/ep-remote/400/400', 'ep-remote.jpg', 0),
	('7af37cf0-88a3-48e8-bb87-aea3a3e36c47', 'image/jpeg', 'https://picsum.photos/seed/ep-bestmovie/400/400', 'ep-bestmovie.jpg', 0),
	('d758437f-1920-43f1-b690-ae4703632baf', 'image/jpeg', 'https://picsum.photos/seed/ep-horror/400/400', 'ep-horror.jpg', 0),
	('d17cf2a2-cdfa-4556-bdbd-e8a06a232edf', 'image/jpeg', 'https://picsum.photos/seed/ep-modernart/400/400', 'ep-modernart.jpg', 0),
	('6fe7ff8c-62c8-4cfe-9f2e-7bed3205b54b', 'image/jpeg', 'https://picsum.photos/seed/ep-worldcup/400/400', 'ep-worldcup.jpg', 0),
	('7e53858c-0052-4559-97a4-82d1061b4dd4', 'image/jpeg', 'https://picsum.photos/seed/ep-marathon/400/400', 'ep-marathon.jpg', 0),
	('12b007e2-dbbb-4d56-85bf-18f7f0d93c56', 'image/jpeg', 'https://picsum.photos/seed/ep-muscle/400/400', 'ep-muscle.jpg', 0),
	('020f1678-90c8-41e5-91a0-6158ca18895c', 'image/jpeg', 'https://picsum.photos/seed/ep-esports/400/400', 'ep-esports.jpg', 0),
	('7a28f39a-14ef-4351-8cdb-5ed5f681db6d', 'image/jpeg', 'https://picsum.photos/seed/ep-baseball/400/400', 'ep-baseball.jpg', 0),
	('686e985c-6150-40e9-aaf6-02eb233e983c', 'image/jpeg', 'https://picsum.photos/seed/ep-olympics/400/400', 'ep-olympics.jpg', 0),
	('9cf4598f-5c99-4fa1-9c0a-9170e9301e2c', 'image/jpeg', 'https://picsum.photos/seed/ep-yoga/400/400', 'ep-yoga.jpg', 0),
	('6f2686c6-b0c4-4013-a3cc-4703ab02c7c5', 'image/jpeg', 'https://picsum.photos/seed/ep-gut/400/400', 'ep-gut.jpg', 0),
	('2fc9b069-d198-4347-8b64-bdf2c9e1cd7e', 'image/jpeg', 'https://picsum.photos/seed/ep-jpop/400/400', 'ep-jpop.jpg', 0),
	('9b2ff367-dc63-48d9-b2a4-4632d4baddf7', 'image/jpeg', 'https://picsum.photos/seed/ep-samurai/400/400', 'ep-samurai.jpg', 0),
	('6b8d6599-655c-4305-b8c6-10ed8d0eb5f4', 'image/jpeg', 'https://picsum.photos/seed/ep-egypt/400/400', 'ep-egypt.jpg', 0),
	('9f01a36e-fa79-48c3-8eb8-c31a899b357e', 'image/jpeg', 'https://picsum.photos/seed/ep-bakumatsu/400/400', 'ep-bakumatsu.jpg', 0),
	('0d26d157-ce7b-42dc-a6bd-033d0d902e31', 'image/jpeg', 'https://picsum.photos/seed/ep-rome/400/400', 'ep-rome.jpg', 0),
	('9aa4bc6c-635f-41ee-9ef7-1ed8b39a1b37', 'image/jpeg', 'https://picsum.photos/seed/ep-castle/400/400', 'ep-castle.jpg', 0),
	('ce424dd4-d8d8-407b-9b8a-655350859e20', 'image/jpeg', 'https://picsum.photos/seed/ep-tonguetwist/400/400', 'ep-tonguetwist.jpg', 0),
	('fff56536-6457-4a09-b4d1-70407c377138', 'image/jpeg', 'https://picsum.photos/seed/ep-aruaru/400/400', 'ep-aruaru.jpg', 0),
	('f457b96d-ff33-4ddd-be36-ad4147da8c31', 'image/jpeg', 'https://picsum.photos/seed/ep-math/400/400', 'ep-math.jpg', 0),
	('53dc6aad-560d-4c48-989c-381187164566', 'image/jpeg', 'https://picsum.photos/seed/ep-starnight/400/400', 'ep-starnight.jpg', 0),
	('2bc1cfe9-a3c2-4542-839b-4c23532385db', 'image/jpeg', 'https://picsum.photos/seed/ep-catmagic/400/400', 'ep-catmagic.jpg', 0),
	('50100001-0006-4000-a000-000000000001', 'image/jpeg', 'https://picsum.photos/seed/ep-solo-diary/400/400', 'ep-solo-diary.jpg', 0);
