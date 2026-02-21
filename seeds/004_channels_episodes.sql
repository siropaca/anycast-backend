-- テスト用のチャンネル・エピソード・台本データを作成する

-- ===========================================
-- チャンネル
-- ===========================================

-- test_user のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('ea9a266e-f532-417c-8916-709d0233941c', '8def69af-dae9-4641-a0e5-100107626933', 'テックトーク', '最新のテクノロジーニュースを2人のパーソナリティが楽しく解説するポッドキャスト', '最新のテクノロジーニュースについて、分かりやすく楽しく解説してください。専門用語は避け、具体例を交えて説明してください。', (SELECT id FROM categories WHERE slug = 'technology'), '6852eda5-db2e-42f0-9870-c00a396e6bf8'),
	('efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', '8def69af-dae9-4641-a0e5-100107626933', 'ゆるふわ雑談ラジオ', '日常のあれこれをゆるく語る雑談番組', 'ゆるい雰囲気で日常の話題について雑談してください。リラックスした会話を心がけてください。', (SELECT id FROM categories WHERE slug = 'society-culture'), 'a28d9fb9-1c8f-4835-a0ad-96afdb6e3279');

-- test_user のソロチャンネル（キャラクター1人）
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('50100001-0001-4000-a000-000000000001', '8def69af-dae9-4641-a0e5-100107626933', 'ひとり語りラジオ', '一人のパーソナリティが日々の気づきや考えをじっくり語るソロポッドキャスト', '日常の出来事や気づきについて、一人でじっくりと語ってください。リスナーに話しかけるような親しみやすい口調で、自分の考えや経験を共有してください。', (SELECT id FROM categories WHERE slug = 'society-culture'), '50100001-0005-4000-a000-000000000001');

-- test_user2 のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('e5a50bd3-8990-4344-b470-56fa7329d75c', '8eada3a5-f413-4eeb-9cd5-12def60d4596', 'ビジネス最前線', '起業やキャリアについて実践的なアドバイスを届けるビジネス番組', '起業やキャリアについて、実践的で具体的なアドバイスを提供してください。成功事例や失敗談を交えて説明してください。', (SELECT id FROM categories WHERE slug = 'business'), 'ef89720a-9e2e-4021-8c28-f4e42d1c3d93');

-- test_user3 (Sakura) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('3908d99b-d52d-4a73-96fb-fca7df2dfec9', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', 'サイエンス・ラボ', '身近な科学をわかりやすく紹介するサイエンス番組', '科学の話題について、身近な例を使ってわかりやすく解説してください。', (SELECT id FROM categories WHERE slug = 'science'), '3cbba00a-e754-4516-bf68-958c84c288e3'),
	('395f3bfa-031e-4d90-a53b-19d311392b00', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', 'ほのぼのライフ', '日常の小さな幸せを見つけるほのぼの番組', 'のんびりした雰囲気で日常の楽しいことについて語ってください。', (SELECT id FROM categories WHERE slug = 'leisure'), '9f17a722-a320-4f6f-a827-05acf96c36c9');

-- test_user4 (Ren) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('7f2c8688-c163-40c3-9303-56b4d8a1ed7a', 'd6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', 'ニュースの裏側', '話題のニュースを深掘りして解説する番組', 'ニュースの背景や裏側を分析して、わかりやすく解説してください。', (SELECT id FROM categories WHERE slug = 'news'), '3084384d-2e85-437a-adfd-a15a0b51414d');

-- test_user5 (Hina) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('7af3be1e-1555-4f71-9af0-2be26cd6b612', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', '映画レビュー倶楽部', '最新映画から名作まで語る映画レビュー番組', '映画について、ネタバレなしで魅力を伝えてください。', (SELECT id FROM categories WHERE slug = 'tv-film'), 'd259698b-7f50-488b-b2c6-5d822be5645d'),
	('892d53ff-0633-4b85-a220-afa0682ee467', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', 'アート散歩', 'アートと文化を気軽に楽しむ番組', 'アートの話題をカジュアルに語ってください。', (SELECT id FROM categories WHERE slug = 'arts'), 'e9c8d1d7-f953-4f3f-a960-6df46ab6eaff');

-- test_user6 (Kaito) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('b77286af-042a-4580-88b6-efe62aaa3eae', '80adf759-b01c-4726-87b4-7b9c659483a4', 'スポーツダイジェスト', 'スポーツの最新情報を熱く語る番組', 'スポーツの話題を熱く、でも分かりやすく伝えてください。', (SELECT id FROM categories WHERE slug = 'sports'), 'f562344f-1bf0-4767-8dd3-34743c92c101');

-- test_user7 (Mio) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('f602c66d-6111-495a-8b9c-e46fa3740d49', '8450d256-8630-4044-8a69-fc8671e6e5c1', 'ヘルシーライフ', '健康とフィットネスの情報を届ける番組', '健康やフィットネスについて実践的なアドバイスをしてください。', (SELECT id FROM categories WHERE slug = 'health-fitness'), 'eb07435e-b994-422f-9d35-f8c7b9e437d3'),
	('bcf2efb2-8a47-46bf-aac4-af593ad0257f', '8450d256-8630-4044-8a69-fc8671e6e5c1', 'ミュージックステーション', '音楽の魅力を語る番組', '音楽について、ジャンルを問わず幅広く語ってください。', (SELECT id FROM categories WHERE slug = 'music'), '90d6d166-3dc0-41f1-bf5e-78c11b25598c');

-- test_user8 (Yuto) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('f02d6a25-4166-4632-812d-446ee6d5be38', '7d4e20d4-98ca-4b79-901d-c51e43c38e2f', '歴史探訪', '歴史の面白エピソードを紹介する番組', '歴史の面白い話を、臨場感たっぷりに語ってください。', (SELECT id FROM categories WHERE slug = 'history'), '1731f558-78bd-4caa-bdc1-aaf7f4875f87');

-- test_user9 (Aoi) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('1737fa8f-75ff-41ba-9775-8d3496d38344', 'c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', 'コメディナイト', '笑いで元気を届けるコメディ番組', '面白おかしく、笑える会話を展開してください。', (SELECT id FROM categories WHERE slug = 'comedy'), 'bb6b9e97-2a5e-4bf0-bc60-814d8cbd99d9'),
	('06891da2-4124-492f-826c-1072dcf27ee6', 'c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', '教育チャンネル', '学びをもっと楽しくする教育番組', '教育的な内容を、楽しく分かりやすく伝えてください。', (SELECT id FROM categories WHERE slug = 'education'), 'ec3b5427-e883-466b-a98d-0f3bc0084da1');

-- test_user10 (Sora) のチャンネル
INSERT INTO channels (id, user_id, name, description, user_prompt, category_id, artwork_id) VALUES
	('70ccbc05-a072-47a2-93c1-e0730d4d2bb9', '767b5ed0-a663-437a-9cc9-b8cef6d0731e', 'フィクション工房', '短編フィクションを朗読する番組', 'オリジナルの短編フィクションを魅力的に語ってください。', (SELECT id FROM categories WHERE slug = 'fiction'), 'bdf2b1f0-735a-489b-be3b-97c97c06a2c8');

-- ===========================================
-- キャラクター（user_id で所有）
-- ===========================================

-- test_user のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('d1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '8def69af-dae9-4641-a0e5-100107626933', 'ユウキ', 'テクノロジーに詳しいエンジニア。論理的だが親しみやすい話し方をする。', (SELECT id FROM voices WHERE name = 'Puck')),
	('4cee85f3-adec-4333-84e6-d6aaefb63408', '8def69af-dae9-4641-a0e5-100107626933', 'ミサキ', '好奇心旺盛なライター。素朴な疑問を投げかけてくれる。', (SELECT id FROM voices WHERE name = 'Zephyr')),
	('b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', '8def69af-dae9-4641-a0e5-100107626933', 'ハルカ', 'のんびり屋でマイペース。ゆるい雰囲気で話を進める。', (SELECT id FROM voices WHERE name = 'Zephyr')),
	('41977119-13d8-4d26-bfe4-694eb2cf2167', '8def69af-dae9-4641-a0e5-100107626933', 'ソウタ', 'ツッコミ担当。ハルカのボケに的確に反応する。', (SELECT id FROM voices WHERE name = 'Puck'));

-- test_user のソロキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('50100001-0002-4000-a000-000000000001', '8def69af-dae9-4641-a0e5-100107626933', 'ナツメ', '落ち着いた語り口の自由人。旅や読書が好きで、日常の小さな発見を大切にする。', (SELECT id FROM voices WHERE name = 'Kore'));

-- test_user2 のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('a4e0f973-f91a-4103-b758-fed371622046', '8eada3a5-f413-4eeb-9cd5-12def60d4596', 'ケンジ', '元外資系コンサルタント。論理的で鋭い視点を持つ。', (SELECT id FROM voices WHERE name = 'Puck')),
	('b7efbbae-0655-46f1-afb7-a42d2646f0c1', '8eada3a5-f413-4eeb-9cd5-12def60d4596', 'アヤカ', 'スタートアップ経営者。実体験に基づいたアドバイスが得意。', (SELECT id FROM voices WHERE name = 'Zephyr'));

-- test_user3 (Sakura) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('d21c4667-77f3-4d80-a36a-3554e310f0dc', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', 'リコ', '理系大学院生。実験大好きな研究者タイプ。', (SELECT id FROM voices WHERE name = 'Aoede')),
	('752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', 'タクミ', 'サイエンスライター。難しいことを簡単に説明するのが得意。', (SELECT id FROM voices WHERE name = 'Charon'));

-- test_user4 (Ren) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('b6442ee5-588e-40ba-8c11-6ae54aa152e9', 'd6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', 'シンイチ', '元新聞記者。鋭い切り口でニュースを読み解く。', (SELECT id FROM voices WHERE name = 'Alnilam')),
	('c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', 'd6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', 'マリ', '政治学者。冷静な分析が持ち味。', (SELECT id FROM voices WHERE name = 'Kore'));

-- test_user5 (Hina) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('41df3d91-9ca7-4aed-ad15-8831d322de5d', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', 'ナオ', '映画評論家。独特の視点で映画を語る。', (SELECT id FROM voices WHERE name = 'Sadachbia')),
	('42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', 'エミ', '映画好きの大学生。素直な感想を伝える。', (SELECT id FROM voices WHERE name = 'Leda'));

-- test_user6 (Kaito) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('c68d60d8-32ef-4484-9dd6-90e487f279fe', '80adf759-b01c-4726-87b4-7b9c659483a4', 'ダイスケ', '元プロサッカー選手。熱血解説が魅力。', (SELECT id FROM voices WHERE name = 'Fenrir')),
	('b5a031d1-c7b2-4a2e-ae8f-301c90c1132b', '80adf759-b01c-4726-87b4-7b9c659483a4', 'サキ', 'スポーツジャーナリスト。データに基づいた分析が得意。', (SELECT id FROM voices WHERE name = 'Achernar'));

-- test_user7 (Mio) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('41a07087-e2f8-46a9-a9d4-213d62964b30', '8450d256-8630-4044-8a69-fc8671e6e5c1', 'カナ', 'ヨガインストラクター。穏やかで前向きな性格。', (SELECT id FROM voices WHERE name = 'Pulcherrima')),
	('825f7a68-0c4d-453a-be8e-2daabef54cc5', '8450d256-8630-4044-8a69-fc8671e6e5c1', 'リュウ', 'パーソナルトレーナー。筋トレ大好き。', (SELECT id FROM voices WHERE name = 'Umbriel'));

-- test_user8 (Yuto) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('f12ff841-52b7-419d-b18f-34b11f0a9b42', '7d4e20d4-98ca-4b79-901d-c51e43c38e2f', 'コウジ', '歴史教授。ユーモアを交えた講義が人気。', (SELECT id FROM voices WHERE name = 'Schedar')),
	('30c345e4-490b-4b30-8796-120e7643bb5a', '7d4e20d4-98ca-4b79-901d-c51e43c38e2f', 'ユイ', '歴史好きの高校生。素朴な疑問を投げかける。', (SELECT id FROM voices WHERE name = 'Despina'));

-- test_user9 (Aoi) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', 'テツヤ', 'お笑い芸人。ボケ担当。', (SELECT id FROM voices WHERE name = 'Iapetus')),
	('b6148fa4-6e9e-4025-a791-b65a327b58e3', 'c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', 'ノゾミ', 'ツッコミ担当。的確なツッコミが光る。', (SELECT id FROM voices WHERE name = 'Sulafat'));

-- test_user10 (Sora) のキャラクター
INSERT INTO characters (id, user_id, name, persona, voice_id) VALUES
	('13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2', '767b5ed0-a663-437a-9cc9-b8cef6d0731e', 'アキラ', '小説家。情感豊かな朗読が魅力。', (SELECT id FROM voices WHERE name = 'Rasalgethi')),
	('5b02ebec-6b83-431a-ae0e-29611d0c7f2f', '767b5ed0-a663-437a-9cc9-b8cef6d0731e', 'ミユ', '声優志望。表現力豊かな演技派。', (SELECT id FROM voices WHERE name = 'Vindemiatrix'));

-- ===========================================
-- チャンネルとキャラクターの紐づけ
-- ===========================================

-- テックトーク（ユウキ、ミサキ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('ea9a266e-f532-417c-8916-709d0233941c', 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007'),
	('ea9a266e-f532-417c-8916-709d0233941c', '4cee85f3-adec-4333-84e6-d6aaefb63408');

-- ゆるふわ雑談ラジオ（ハルカ、ソウタ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', 'b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4'),
	('efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', '41977119-13d8-4d26-bfe4-694eb2cf2167');

-- ひとり語りラジオ（ナツメ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('50100001-0001-4000-a000-000000000001', '50100001-0002-4000-a000-000000000001');

-- ビジネス最前線（ケンジ、アヤカ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('e5a50bd3-8990-4344-b470-56fa7329d75c', 'a4e0f973-f91a-4103-b758-fed371622046'),
	('e5a50bd3-8990-4344-b470-56fa7329d75c', 'b7efbbae-0655-46f1-afb7-a42d2646f0c1');

-- サイエンス・ラボ（リコ、タクミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('3908d99b-d52d-4a73-96fb-fca7df2dfec9', 'd21c4667-77f3-4d80-a36a-3554e310f0dc'),
	('3908d99b-d52d-4a73-96fb-fca7df2dfec9', '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa');

-- ほのぼのライフ（リコ、タクミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('395f3bfa-031e-4d90-a53b-19d311392b00', 'd21c4667-77f3-4d80-a36a-3554e310f0dc'),
	('395f3bfa-031e-4d90-a53b-19d311392b00', '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa');

-- ニュースの裏側（シンイチ、マリ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('7f2c8688-c163-40c3-9303-56b4d8a1ed7a', 'b6442ee5-588e-40ba-8c11-6ae54aa152e9'),
	('7f2c8688-c163-40c3-9303-56b4d8a1ed7a', 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40');

-- 映画レビュー倶楽部（ナオ、エミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('7af3be1e-1555-4f71-9af0-2be26cd6b612', '41df3d91-9ca7-4aed-ad15-8831d322de5d'),
	('7af3be1e-1555-4f71-9af0-2be26cd6b612', '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7');

-- アート散歩（ナオ、エミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('892d53ff-0633-4b85-a220-afa0682ee467', '41df3d91-9ca7-4aed-ad15-8831d322de5d'),
	('892d53ff-0633-4b85-a220-afa0682ee467', '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7');

-- スポーツダイジェスト（ダイスケ、サキ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('b77286af-042a-4580-88b6-efe62aaa3eae', 'c68d60d8-32ef-4484-9dd6-90e487f279fe'),
	('b77286af-042a-4580-88b6-efe62aaa3eae', 'b5a031d1-c7b2-4a2e-ae8f-301c90c1132b');

-- ヘルシーライフ（カナ、リュウ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('f602c66d-6111-495a-8b9c-e46fa3740d49', '41a07087-e2f8-46a9-a9d4-213d62964b30'),
	('f602c66d-6111-495a-8b9c-e46fa3740d49', '825f7a68-0c4d-453a-be8e-2daabef54cc5');

-- ミュージックステーション（カナ、リュウ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('bcf2efb2-8a47-46bf-aac4-af593ad0257f', '41a07087-e2f8-46a9-a9d4-213d62964b30'),
	('bcf2efb2-8a47-46bf-aac4-af593ad0257f', '825f7a68-0c4d-453a-be8e-2daabef54cc5');

-- 歴史探訪（コウジ、ユイ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('f02d6a25-4166-4632-812d-446ee6d5be38', 'f12ff841-52b7-419d-b18f-34b11f0a9b42'),
	('f02d6a25-4166-4632-812d-446ee6d5be38', '30c345e4-490b-4b30-8796-120e7643bb5a');

-- コメディナイト（テツヤ、ノゾミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('1737fa8f-75ff-41ba-9775-8d3496d38344', 'ef1c418c-8c10-4b02-887b-2ab1ae937f56'),
	('1737fa8f-75ff-41ba-9775-8d3496d38344', 'b6148fa4-6e9e-4025-a791-b65a327b58e3');

-- 教育チャンネル（テツヤ、ノゾミ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('06891da2-4124-492f-826c-1072dcf27ee6', 'ef1c418c-8c10-4b02-887b-2ab1ae937f56'),
	('06891da2-4124-492f-826c-1072dcf27ee6', 'b6148fa4-6e9e-4025-a791-b65a327b58e3');

-- フィクション工房（アキラ、ミユ）
INSERT INTO channel_characters (channel_id, character_id) VALUES
	('70ccbc05-a072-47a2-93c1-e0730d4d2bb9', '13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2'),
	('70ccbc05-a072-47a2-93c1-e0730d4d2bb9', '5b02ebec-6b83-431a-ae0e-29611d0c7f2f');

-- ===========================================
-- エピソード
-- ===========================================

-- test_user のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('eb960304-f86e-4364-be5d-d3d5126c9601', 'ea9a266e-f532-417c-8916-709d0233941c', 'AI の未来を語る', 'ChatGPT から始まった AI ブームの今後について', NOW(), '209e8d96-dbc1-4dcf-a747-a0ac4ec37a0c'),
	('67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 'ea9a266e-f532-417c-8916-709d0233941c', 'スマートホームのすすめ', '自宅を便利にするガジェット紹介', NOW(), 'e463cd78-0e09-4d6b-ac83-ec01fb25d5a9'),
	('198d7e19-7d40-4299-95bf-a641f5c83911', 'efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', '最近ハマってること', 'お互いの趣味について語り合う回', NULL, '783ac817-5deb-4711-a8d1-b3e9721b891c'),
	('a0100001-0001-4000-a000-000000000001', 'ea9a266e-f532-417c-8916-709d0233941c', 'プログラミング言語の選び方', '初心者から上級者まで、目的に合ったプログラミング言語の選び方を徹底討論', NULL, 'a0100001-0002-4000-a000-000000000001');

-- test_user のソロエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('50100001-0003-4000-a000-000000000001', '50100001-0001-4000-a000-000000000001', '朝の散歩で見つけたこと', '毎朝の散歩で気づいた季節の移ろいと、日常の小さな幸せについて', NOW(), '50100001-0006-4000-a000-000000000001');

-- test_user2 のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('fcb16526-951a-4ff1-a456-ab1dba96f699', 'e5a50bd3-8990-4344-b470-56fa7329d75c', '副業から始める起業入門', 'リスクを抑えながら起業にチャレンジする方法', NOW(), '45886285-2b5b-4575-8c5c-9c95d139a803'),
	('9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 'e5a50bd3-8990-4344-b470-56fa7329d75c', '失敗しない資金調達の秘訣', 'スタートアップの資金調達で気をつけるべきポイント', NOW(), '2a2dacc8-670e-4d3d-965c-9c8ae6679bd5');

-- test_user3 (Sakura) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('b1c1e7d7-b3eb-4783-82d0-6857832daf09', '3908d99b-d52d-4a73-96fb-fca7df2dfec9', '量子コンピュータ入門', '量子コンピュータの仕組みをゼロから解説', NOW(), 'aae14562-a3a1-4507-8f24-bb72a11dda5c'),
	('682c05ed-3ae8-4558-a74c-62f8cdc7b344', '3908d99b-d52d-4a73-96fb-fca7df2dfec9', '宇宙の神秘に迫る', 'ブラックホールやダークマターの最新研究', NOW(), '9a7dc980-2da8-4bf1-9557-9a52193af39b'),
	('1a4aad00-fd65-4960-a11f-45f2b5ff504c', '3908d99b-d52d-4a73-96fb-fca7df2dfec9', 'DNA と遺伝子の不思議', '遺伝子編集技術の現在と未来', NOW(), 'ccaedc4a-59bc-4736-980e-699d85dd822b'),
	('98f515d1-ca1f-4810-b8b6-a147aac641f4', '395f3bfa-031e-4d90-a53b-19d311392b00', 'お気に入りのカフェ巡り', '街の隠れた名店を紹介する回', NOW(), '21188860-1bc2-45f2-a15c-10462e2c20cc'),
	('29c31347-3078-4ab0-9773-f6d408462c39', '395f3bfa-031e-4d90-a53b-19d311392b00', '休日の過ごし方', '理想の休日について語る回', NULL, '8b53c1a7-a536-4acb-8a61-aaf6eea880db');

-- test_user4 (Ren) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('436043d0-9a74-4541-9639-6b9566125bd7', '7f2c8688-c163-40c3-9303-56b4d8a1ed7a', 'SNS 時代のメディアリテラシー', 'フェイクニュースにどう向き合うか', NOW(), '679e7f0d-ca8a-46ca-b7c5-a4c44ddecc28'),
	('c4d39d21-4e39-4150-bd3d-5ceb27718a38', '7f2c8688-c163-40c3-9303-56b4d8a1ed7a', '選挙と民主主義の未来', '投票率低下の問題を考える', NOW(), '467ae4ef-5302-4750-b71e-12a74d5c9860'),
	('8cfcfff0-87c8-4db9-a250-7ccc5fba1407', '7f2c8688-c163-40c3-9303-56b4d8a1ed7a', '気候変動と私たちの暮らし', '身近にできる環境対策を考える', NOW(), '1971d88a-91f5-4b72-9e0b-9319fe0c3e1a'),
	('7c2d9c7a-e3e3-4388-bfe7-d36162a19768', '7f2c8688-c163-40c3-9303-56b4d8a1ed7a', 'リモートワーク革命', '働き方の変化を追う', NOW(), '4010aa53-c71b-4640-94c2-283cf3ff0a02');

-- test_user5 (Hina) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('b7a74a63-a9e2-4f60-ba58-bfa887598e07', '7af3be1e-1555-4f71-9af0-2be26cd6b612', '今年のベスト映画 TOP5', '年間ベスト映画を発表', NOW(), '7af37cf0-88a3-48e8-bb87-aea3a3e36c47'),
	('105d9802-0133-4a06-a1b4-de87eb26f5ac', '7af3be1e-1555-4f71-9af0-2be26cd6b612', 'ホラー映画の魅力', 'なぜ人はホラーに惹かれるのか', NOW(), 'd758437f-1920-43f1-b690-ae4703632baf'),
	('ec47dd8e-34c5-4ba4-a101-02713e87ba74', '892d53ff-0633-4b85-a220-afa0682ee467', '現代アートの楽しみ方', '難しくない！現代アート入門', NOW(), 'd17cf2a2-cdfa-4556-bdbd-e8a06a232edf');

-- test_user6 (Kaito) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a', 'b77286af-042a-4580-88b6-efe62aaa3eae', 'サッカー W 杯を振り返る', 'W 杯の名場面を語る', NOW(), '6fe7ff8c-62c8-4cfe-9f2e-7bed3205b54b'),
	('ecd6cb92-597d-4367-9c65-9f98fb38fc9b', 'b77286af-042a-4580-88b6-efe62aaa3eae', 'マラソンの科学', '市民ランナーのための科学的トレーニング', NOW(), '7e53858c-0052-4559-97a4-82d1061b4dd4'),
	('f270e88b-4c7f-4fd5-a06d-5a52ab0ff501', 'b77286af-042a-4580-88b6-efe62aaa3eae', '筋トレとメンタルの関係', '運動がメンタルヘルスに与える影響', NOW(), '12b007e2-dbbb-4d56-85bf-18f7f0d93c56'),
	('80d5ac1d-4a0c-44b9-8e4c-61b2ac333b00', 'b77286af-042a-4580-88b6-efe62aaa3eae', 'eスポーツの今', 'eスポーツは本当のスポーツか？', NOW(), '020f1678-90c8-41e5-91a0-6158ca18895c'),
	('948ee16d-ef1b-4e54-8dcb-cf6b7b46c7fd', 'b77286af-042a-4580-88b6-efe62aaa3eae', '野球データ分析入門', 'セイバーメトリクスで野球を楽しむ', NOW(), '7a28f39a-14ef-4351-8cdb-5ed5f681db6d'),
	('4288967f-f224-452a-bf4d-f0c61a1ecaa9', 'b77286af-042a-4580-88b6-efe62aaa3eae', 'オリンピックの感動秘話', 'オリンピック選手のドラマチックなエピソード', NOW(), '686e985c-6150-40e9-aaf6-02eb233e983c');

-- test_user7 (Mio) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('9890df9b-05cd-4962-9c7f-42e05b5c6ced', 'f602c66d-6111-495a-8b9c-e46fa3740d49', '朝ヨガのすすめ', '朝の 10 分ヨガで一日を変える', NOW(), '9cf4598f-5c99-4fa1-9c0a-9170e9301e2c'),
	('fa2a8e34-234b-4d40-acad-e4448d1476d0', 'f602c66d-6111-495a-8b9c-e46fa3740d49', '腸活で健康に', '腸内環境を整える食事法', NOW(), '6f2686c6-b0c4-4013-a3cc-4703ab02c7c5'),
	('a787628e-8dd9-4cac-853d-8d84b0aec6ec', 'bcf2efb2-8a47-46bf-aac4-af593ad0257f', 'J-POP の歴史を辿る', '90 年代から現在までの J-POP の変遷', NOW(), '2fc9b069-d198-4347-8b64-bdf2c9e1cd7e');

-- test_user8 (Yuto) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('2267a065-9958-493b-9cb3-c62e578b7c23', 'f02d6a25-4166-4632-812d-446ee6d5be38', '戦国武将の意外な一面', '教科書には載らない武将エピソード', NOW(), '9b2ff367-dc63-48d9-b2a4-4632d4baddf7'),
	('3c8a7f95-0001-4437-a1ec-138009cd0001', 'f02d6a25-4166-4632-812d-446ee6d5be38', '古代エジプトの謎', 'ピラミッドの建設方法を考える', NOW(), '6b8d6599-655c-4305-b8c6-10ed8d0eb5f4'),
	('3c8a7f95-0002-4437-a1ec-138009cd0002', 'f02d6a25-4166-4632-812d-446ee6d5be38', '幕末の志士たち', '明治維新を支えた若者たちの物語', NOW(), '9f01a36e-fa79-48c3-8eb8-c31a899b357e'),
	('3c8a7f95-0003-4437-a1ec-138009cd0003', 'f02d6a25-4166-4632-812d-446ee6d5be38', 'ローマ帝国の栄光と衰退', '世界史最大の帝国の物語', NOW(), '0d26d157-ce7b-42dc-a6bd-033d0d902e31'),
	('3c8a7f95-0004-4437-a1ec-138009cd0004', 'f02d6a25-4166-4632-812d-446ee6d5be38', '日本の城の秘密', 'お城に隠された建築の知恵', NOW(), '9aa4bc6c-635f-41ee-9ef7-1ed8b39a1b37');

-- test_user9 (Aoi) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('3c8a7f95-0005-4437-a1ec-138009cd0005', '1737fa8f-75ff-41ba-9775-8d3496d38344', '笑ってはいけない早口言葉', '早口言葉チャレンジで爆笑', NOW(), 'ce424dd4-d8d8-407b-9b8a-655350859e20'),
	('3c8a7f95-0006-4437-a1ec-138009cd0006', '1737fa8f-75ff-41ba-9775-8d3496d38344', 'あるあるネタ大会', '日常のあるあるで盛り上がる回', NOW(), 'fff56536-6457-4a09-b4d1-70407c377138'),
	('3c8a7f95-0007-4437-a1ec-138009cd0007', '06891da2-4124-492f-826c-1072dcf27ee6', '数学を好きになる方法', '苦手意識を克服するコツ', NOW(), 'f457b96d-ff33-4ddd-be36-ad4147da8c31');

-- test_user10 (Sora) のエピソード
INSERT INTO episodes (id, channel_id, title, description, published_at, artwork_id) VALUES
	('3c8a7f95-0008-4437-a1ec-138009cd0008', '70ccbc05-a072-47a2-93c1-e0730d4d2bb9', '星降る夜の物語', 'ファンタジー短編：星の世界への冒険', NOW(), '53dc6aad-560d-4c48-989c-381187164566'),
	('3c8a7f95-0009-4437-a1ec-138009cd0009', '70ccbc05-a072-47a2-93c1-e0730d4d2bb9', '猫と魔法使い', 'ファンタジー短編：猫が魔法を使う物語', NOW(), '2bc1cfe9-a3c2-4542-839b-4c23532385db');

-- ===========================================
-- 台本（ScriptLines）
-- ===========================================

-- Episode 1: AI の未来を語る
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('236f9071-900a-4b75-aea7-ebb847f5ccad', 'eb960304-f86e-4364-be5d-d3d5126c9601', 0, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'こんにちは、テックトークへようこそ！今日は AI の未来について話していきます。', NULL),
	('bff9d166-1ad5-46fa-96fb-39a27378e99f', 'eb960304-f86e-4364-be5d-d3d5126c9601', 1, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'よろしくお願いします！最近 ChatGPT がすごく話題ですよね。', NULL),
	('d5422671-73c8-4b28-afe1-5b0c419dcd49', 'eb960304-f86e-4364-be5d-d3d5126c9601', 2, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'そうなんです。大規模言語モデルの進化は目覚ましいものがあります。', NULL),
	('51223f66-3ac5-4685-9609-50d0ccd9b10a', 'eb960304-f86e-4364-be5d-d3d5126c9601', 3, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'いつか AI がポッドキャストを作る時代が来るかもしれませんね！', '笑いながら');

-- Episode 2: スマートホームのすすめ
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('d8909d91-da04-4ec6-bec1-356eb9c4e2d9', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 0, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '今日はスマートホームについて紹介していきます。', NULL),
	('f0d41215-6172-4bca-a10e-efaa002a09fc', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 1, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'スマートホームって難しそうなイメージがあるんですけど', NULL),
	('6723d570-d7c6-4a07-b481-0b609765be86', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 2, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '私も最初は全然分からなかったんですよ。でもやってみたら意外と簡単でした。', '笑いながら'),
	('504ebb03-a05c-49ce-9e94-974f9cc80cc0', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 3, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'なるほど！それなら私でもできそうです。', NULL);

-- Episode 3: 最近ハマってること
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('23e48682-4949-4751-aef0-b80e369a899b', '198d7e19-7d40-4299-95bf-a641f5c83911', 0, 'b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', 'ねえねえ、最近なんかハマってることある？', NULL),
	('8b7f30af-1662-413d-84b6-2b27033435f7', '198d7e19-7d40-4299-95bf-a641f5c83911', 1, '41977119-13d8-4d26-bfe4-694eb2cf2167', '最近はコーヒーにハマってるかな。豆から挽いて淹れてるよ。', NULL),
	('fd9ce404-74c8-456e-8f1b-eda25e22ccce', '198d7e19-7d40-4299-95bf-a641f5c83911', 2, 'b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', 'すごい凝ってるね！私なんてインスタントで十分だよ。', '笑いながら'),
	('18641a58-561f-466a-a670-cf3a569c6669', '198d7e19-7d40-4299-95bf-a641f5c83911', 3, '41977119-13d8-4d26-bfe4-694eb2cf2167', 'インスタントも美味しいよね。手軽さって大事。', NULL);

-- Episode: 朝の散歩で見つけたこと（test_user / ひとり語りラジオ / ソロ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('50100001-0004-4000-a000-000000000001', '50100001-0003-4000-a000-000000000001', 0, '50100001-0002-4000-a000-000000000001', 'おはようございます、ひとり語りラジオのナツメです。今日もゆるっとお話ししていきますね。', NULL),
	('50100001-0004-4000-a000-000000000002', '50100001-0003-4000-a000-000000000001', 1, '50100001-0002-4000-a000-000000000001', '最近ね、毎朝の散歩を始めたんですよ。朝6時くらいに起きて、近所の公園をぐるっと一周するんです。', NULL),
	('50100001-0004-4000-a000-000000000003', '50100001-0003-4000-a000-000000000001', 2, '50100001-0002-4000-a000-000000000001', 'そしたらね、毎日同じ道を歩いてるのに、少しずつ景色が変わっていくのに気づいて。', NULL),
	('50100001-0004-4000-a000-000000000004', '50100001-0003-4000-a000-000000000001', 3, '50100001-0002-4000-a000-000000000001', 'こういう小さな変化に気づけるようになったのが、なんだかすごく嬉しくて。皆さんもぜひ試してみてくださいね。', '穏やかに');

-- Episode 4: 副業から始める起業入門（test_user2）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('ae5f21f0-a737-47cf-8d00-e9f490bea753', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 0, 'a4e0f973-f91a-4103-b758-fed371622046', '今日は副業から起業を始める方法についてお話しします。', NULL),
	('1a87b77a-2211-4421-9f2c-334ce913e5c3', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 1, 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '私も最初は副業からスタートしました。いきなり会社を辞めるのはリスクが高いですからね。', NULL),
	('b8f262e5-f027-484a-a0f5-997e5b9dd569', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 2, 'a4e0f973-f91a-4103-b758-fed371622046', '私なんて最初の副業で赤字出しちゃいましたからね。今となっては良い思い出ですけど。', '笑いながら'),
	('da61ccde-dfea-4ad8-8a84-f8c4d5a79ac3', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 3, 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', 'そうですね。失敗しても学びになりますし、挑戦することが大切です。', NULL);

-- Episode 5: 失敗しない資金調達の秘訣（test_user2）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('089e59a2-e26b-4dcc-aeca-6763a7ab16b9', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 0, 'a4e0f973-f91a-4103-b758-fed371622046', '今回は資金調達について詳しくお話ししていきます。', NULL),
	('83fccc88-5647-47df-9757-5ebd19b301c7', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 1, 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '資金調達って種類がたくさんありますよね。VC、エンジェル投資家、融資', NULL),
	('8f197ebe-677f-4d9d-add7-111af58b6c04', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 2, 'a4e0f973-f91a-4103-b758-fed371622046', 'その通りです。事業のフェーズによって最適な調達方法は変わってきます。', NULL),
	('2fecb517-5d29-4f74-b3a7-7a85700e4e22', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 3, 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '私も最初のピッチでは緊張しすぎて、投資家の名前を間違えちゃいました。', '笑いながら');

-- Episode 6: 量子コンピュータ入門（test_user3 / サイエンス・ラボ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000001-0001-4000-a000-000000000001', 'b1c1e7d7-b3eb-4783-82d0-6857832daf09', 0, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', 'さて、今日は量子コンピュータについて話していくよ！', NULL),
	('a0000001-0001-4000-a000-000000000002', 'b1c1e7d7-b3eb-4783-82d0-6857832daf09', 1, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', '量子コンピュータって名前はよく聞くけど、結局何がすごいの？', NULL),
	('a0000001-0001-4000-a000-000000000003', 'b1c1e7d7-b3eb-4783-82d0-6857832daf09', 2, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', '簡単に言うと、従来のコンピュータでは何万年もかかる計算を一瞬で解けるんだ。', NULL),
	('a0000001-0001-4000-a000-000000000004', 'b1c1e7d7-b3eb-4783-82d0-6857832daf09', 3, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', 'それは革命的だね！でもまだ実用化は先なのかな？', NULL);

-- Episode 7: 宇宙の神秘に迫る（test_user3 / サイエンス・ラボ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000002-0001-4000-a000-000000000001', '682c05ed-3ae8-4558-a74c-62f8cdc7b344', 0, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', '今日は宇宙の話をしよう！ブラックホールって知ってる？', NULL),
	('a0000002-0001-4000-a000-000000000002', '682c05ed-3ae8-4558-a74c-62f8cdc7b344', 1, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', 'もちろん！光すら脱出できないやつだよね。', NULL),
	('a0000002-0001-4000-a000-000000000003', '682c05ed-3ae8-4558-a74c-62f8cdc7b344', 2, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', 'そう！最近の研究で、ブラックホールの中の構造が少しずつ分かってきたんだ。', NULL),
	('a0000002-0001-4000-a000-000000000004', '682c05ed-3ae8-4558-a74c-62f8cdc7b344', 3, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', 'えっ、中の構造って！？めちゃくちゃ気になる！', '驚きながら');

-- Episode 8: DNA と遺伝子の不思議（test_user3 / サイエンス・ラボ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000003-0001-4000-a000-000000000001', '1a4aad00-fd65-4960-a11f-45f2b5ff504c', 0, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', '今回は遺伝子編集技術、CRISPR について話していきましょう。', NULL),
	('a0000003-0001-4000-a000-000000000002', '1a4aad00-fd65-4960-a11f-45f2b5ff504c', 1, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', 'CRISPR は DNA をハサミのように切り貼りできる技術なんだよね。', NULL),
	('a0000003-0001-4000-a000-000000000003', '1a4aad00-fd65-4960-a11f-45f2b5ff504c', 2, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', '医療への応用が期待されているけど、倫理的な問題もあるよね。', NULL),
	('a0000003-0001-4000-a000-000000000004', '1a4aad00-fd65-4960-a11f-45f2b5ff504c', 3, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', 'そこが一番大事なポイントだね。技術と倫理のバランスが問われている。', NULL);

-- Episode 9: お気に入りのカフェ巡り（test_user3 / ほのぼのライフ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000004-0001-4000-a000-000000000001', '98f515d1-ca1f-4810-b8b6-a147aac641f4', 0, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', 'ねぇねぇ、最近いいカフェ見つけたんだ！', NULL),
	('a0000004-0001-4000-a000-000000000002', '98f515d1-ca1f-4810-b8b6-a147aac641f4', 1, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', 'お、どんなところ？', NULL),
	('a0000004-0001-4000-a000-000000000003', '98f515d1-ca1f-4810-b8b6-a147aac641f4', 2, 'd21c4667-77f3-4d80-a36a-3554e310f0dc', '路地裏にある小さなお店でね、自家焙煎のコーヒーが絶品なの！', '嬉しそうに'),
	('a0000004-0001-4000-a000-000000000004', '98f515d1-ca1f-4810-b8b6-a147aac641f4', 3, '752a32c9-bdbf-42fb-adb4-e3d4ff7080fa', 'いいなぁ、今度一緒に行こうよ。', NULL);

-- Episode 10: SNS 時代のメディアリテラシー（test_user4 / ニュースの裏側）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000005-0001-4000-a000-000000000001', '436043d0-9a74-4541-9639-6b9566125bd7', 0, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', 'SNS の情報をどこまで信じるべきか、今日はこのテーマで話します。', NULL),
	('a0000005-0001-4000-a000-000000000002', '436043d0-9a74-4541-9639-6b9566125bd7', 1, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', 'フェイクニュースの問題は年々深刻になっていますよね。', NULL),
	('a0000005-0001-4000-a000-000000000003', '436043d0-9a74-4541-9639-6b9566125bd7', 2, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', '特に選挙期間中は注意が必要です。情報の出典を確認する習慣が大切ですね。', NULL),
	('a0000005-0001-4000-a000-000000000004', '436043d0-9a74-4541-9639-6b9566125bd7', 3, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', 'ファクトチェックの方法も具体的に紹介していきましょう。', NULL);

-- Episode 11: 選挙と民主主義の未来（test_user4 / ニュースの裏側）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000006-0001-4000-a000-000000000001', 'c4d39d21-4e39-4150-bd3d-5ceb27718a38', 0, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', '投票率の低下が続いていますが、これは民主主義の危機でしょうか。', NULL),
	('a0000006-0001-4000-a000-000000000002', 'c4d39d21-4e39-4150-bd3d-5ceb27718a38', 1, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', '若者の政治離れとよく言われますが、実はそう単純な話ではないんです。', NULL),
	('a0000006-0001-4000-a000-000000000003', 'c4d39d21-4e39-4150-bd3d-5ceb27718a38', 2, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', 'オンライン投票の導入なども議論されていますよね。', NULL),
	('a0000006-0001-4000-a000-000000000004', 'c4d39d21-4e39-4150-bd3d-5ceb27718a38', 3, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', 'テクノロジーと民主主義の関係は今後ますます重要になっていきます。', NULL);

-- Episode 12: 気候変動と私たちの暮らし（test_user4 / ニュースの裏側）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000007-0001-4000-a000-000000000001', '8cfcfff0-87c8-4db9-a250-7ccc5fba1407', 0, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', '気候変動が身近な生活にどう影響しているか、考えてみましょう。', NULL),
	('a0000007-0001-4000-a000-000000000002', '8cfcfff0-87c8-4db9-a250-7ccc5fba1407', 1, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', '猛暑日の増加や豪雨の頻発化は確実に増えていますね。', NULL),
	('a0000007-0001-4000-a000-000000000003', '8cfcfff0-87c8-4db9-a250-7ccc5fba1407', 2, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', '個人でできることも実はたくさんあるんですよ。', NULL),
	('a0000007-0001-4000-a000-000000000004', '8cfcfff0-87c8-4db9-a250-7ccc5fba1407', 3, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', '具体的なアクションを紹介していきましょう。', NULL);

-- Episode 13: リモートワーク革命（test_user4 / ニュースの裏側）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000008-0001-4000-a000-000000000001', '7c2d9c7a-e3e3-4388-bfe7-d36162a19768', 0, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', 'コロナ以降、働き方は大きく変わりましたね。', NULL),
	('a0000008-0001-4000-a000-000000000002', '7c2d9c7a-e3e3-4388-bfe7-d36162a19768', 1, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', 'リモートワークが定着した企業と、出社に戻した企業に分かれていますね。', NULL),
	('a0000008-0001-4000-a000-000000000003', '7c2d9c7a-e3e3-4388-bfe7-d36162a19768', 2, 'c2c3ceb2-4fe1-407f-8ed2-91ae7eb02b40', '生産性の観点からはどうなんでしょう？', NULL),
	('a0000008-0001-4000-a000-000000000004', '7c2d9c7a-e3e3-4388-bfe7-d36162a19768', 3, 'b6442ee5-588e-40ba-8c11-6ae54aa152e9', '調査データを見ると、実は職種によって大きく異なるんです。', NULL);

-- Episode 14: 今年のベスト映画 TOP5（test_user5 / 映画レビュー倶楽部）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000009-0001-4000-a000-000000000001', 'b7a74a63-a9e2-4f60-ba58-bfa887598e07', 0, '41df3d91-9ca7-4aed-ad15-8831d322de5d', 'さあ、年末恒例のベスト映画発表です！', NULL),
	('a0000009-0001-4000-a000-000000000002', 'b7a74a63-a9e2-4f60-ba58-bfa887598e07', 1, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', '今年は豊作でしたよね！選ぶのが大変でした。', NULL),
	('a0000009-0001-4000-a000-000000000003', 'b7a74a63-a9e2-4f60-ba58-bfa887598e07', 2, '41df3d91-9ca7-4aed-ad15-8831d322de5d', '第5位から順番に発表していきましょう！', 'ワクワクしながら'),
	('a0000009-0001-4000-a000-000000000004', 'b7a74a63-a9e2-4f60-ba58-bfa887598e07', 3, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', '楽しみ！私のランキングと比べてみたいです。', NULL);

-- Episode 15: ホラー映画の魅力（test_user5 / 映画レビュー倶楽部）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000010-0001-4000-a000-000000000001', '105d9802-0133-4a06-a1b4-de87eb26f5ac', 0, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', '今日はホラー映画特集！私、実は苦手なんですけど', '少し怖がりながら'),
	('a0000010-0001-4000-a000-000000000002', '105d9802-0133-4a06-a1b4-de87eb26f5ac', 1, '41df3d91-9ca7-4aed-ad15-8831d322de5d', 'ホラーには人間の本質に迫る深いテーマが隠れているんだよ。', NULL),
	('a0000010-0001-4000-a000-000000000003', '105d9802-0133-4a06-a1b4-de87eb26f5ac', 2, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', 'そう言われると確かに、社会風刺的な作品も多いですよね。', NULL),
	('a0000010-0001-4000-a000-000000000004', '105d9802-0133-4a06-a1b4-de87eb26f5ac', 3, '41df3d91-9ca7-4aed-ad15-8831d322de5d', 'そう！ジョーダン・ピールの作品なんかまさにそうだよね。', NULL);

-- Episode 16: 現代アートの楽しみ方（test_user5 / アート散歩）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000011-0001-4000-a000-000000000001', 'ec47dd8e-34c5-4ba4-a101-02713e87ba74', 0, '41df3d91-9ca7-4aed-ad15-8831d322de5d', '現代アートって「わからない」って言われがちだけど、そこがいいんだよ。', NULL),
	('a0000011-0001-4000-a000-000000000002', 'ec47dd8e-34c5-4ba4-a101-02713e87ba74', 1, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', 'え、わからなくていいんですか？', NULL),
	('a0000011-0001-4000-a000-000000000003', 'ec47dd8e-34c5-4ba4-a101-02713e87ba74', 2, '41df3d91-9ca7-4aed-ad15-8831d322de5d', '自分なりの解釈を楽しむのが現代アートの醍醐味なんだ。', NULL),
	('a0000011-0001-4000-a000-000000000004', 'ec47dd8e-34c5-4ba4-a101-02713e87ba74', 3, '42209c1d-4ca5-4dd9-89fc-8c92bb468bb7', 'なるほど、自由に感じていいんだ。気が楽になりました！', '笑いながら');

-- Episode 17: サッカー W 杯を振り返る（test_user6 / スポーツダイジェスト）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000012-0001-4000-a000-000000000001', '4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a', 0, 'c68d60d8-32ef-4484-9dd6-90e487f279fe', 'W 杯の名場面を振り返ろう！まずはあのゴールからだ！', NULL),
	('a0000012-0001-4000-a000-000000000002', '4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a', 1, 'b5a031d1-c7b2-4a2e-ae8f-301c90c1132b', 'データで見ると、あの試合のボール支配率は驚異的でしたね。', NULL),
	('a0000012-0001-4000-a000-000000000003', '4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a', 2, 'c68d60d8-32ef-4484-9dd6-90e487f279fe', 'でもサッカーはデータだけじゃない！あの熱狂を忘れられへんわ！', '興奮しながら'),
	('a0000012-0001-4000-a000-000000000004', '4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a', 3, 'b5a031d1-c7b2-4a2e-ae8f-301c90c1132b', 'たしかに、スポーツの魅力はデータを超えたところにありますね。', NULL);

-- Episode 18: マラソンの科学（test_user6 / スポーツダイジェスト）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000013-0001-4000-a000-000000000001', 'ecd6cb92-597d-4367-9c65-9f98fb38fc9b', 0, 'b5a031d1-c7b2-4a2e-ae8f-301c90c1132b', '今日は市民ランナー向けの科学的トレーニング法を紹介します。', NULL),
	('a0000013-0001-4000-a000-000000000002', 'ecd6cb92-597d-4367-9c65-9f98fb38fc9b', 1, 'c68d60d8-32ef-4484-9dd6-90e487f279fe', 'ワイも現役時代に走り込みはめっちゃやったで！', NULL),
	('a0000013-0001-4000-a000-000000000003', 'ecd6cb92-597d-4367-9c65-9f98fb38fc9b', 2, 'b5a031d1-c7b2-4a2e-ae8f-301c90c1132b', '心拍数ゾーンを意識したトレーニングが効果的なんです。', NULL),
	('a0000013-0001-4000-a000-000000000004', 'ecd6cb92-597d-4367-9c65-9f98fb38fc9b', 3, 'c68d60d8-32ef-4484-9dd6-90e487f279fe', 'なるほど、がむしゃらに走るだけじゃダメなんやな。', NULL);

-- Episode 19: 朝ヨガのすすめ（test_user7 / ヘルシーライフ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000014-0001-4000-a000-000000000001', '9890df9b-05cd-4962-9c7f-42e05b5c6ced', 0, '41a07087-e2f8-46a9-a9d4-213d62964b30', '朝の 10 分で体も心もスッキリするヨガを紹介します。', NULL),
	('a0000014-0001-4000-a000-000000000002', '9890df9b-05cd-4962-9c7f-42e05b5c6ced', 1, '825f7a68-0c4d-453a-be8e-2daabef54cc5', '朝にストレッチするだけでも全然違うよね。', NULL),
	('a0000014-0001-4000-a000-000000000003', '9890df9b-05cd-4962-9c7f-42e05b5c6ced', 2, '41a07087-e2f8-46a9-a9d4-213d62964b30', '太陽礼拝から始めると、一日のエネルギーが全然変わりますよ。', NULL),
	('a0000014-0001-4000-a000-000000000004', '9890df9b-05cd-4962-9c7f-42e05b5c6ced', 3, '825f7a68-0c4d-453a-be8e-2daabef54cc5', '俺も筋トレ前にヨガ取り入れてみようかな。', NULL);

-- Episode 20: 腸活で健康に（test_user7 / ヘルシーライフ）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000015-0001-4000-a000-000000000001', 'fa2a8e34-234b-4d40-acad-e4448d1476d0', 0, '825f7a68-0c4d-453a-be8e-2daabef54cc5', '最近「腸活」って流行ってるけど、実際どうなの？', NULL),
	('a0000015-0001-4000-a000-000000000002', 'fa2a8e34-234b-4d40-acad-e4448d1476d0', 1, '41a07087-e2f8-46a9-a9d4-213d62964b30', '腸は第二の脳とも言われていて、免疫にもメンタルにも関わっているんです。', NULL),
	('a0000015-0001-4000-a000-000000000003', 'fa2a8e34-234b-4d40-acad-e4448d1476d0', 2, '825f7a68-0c4d-453a-be8e-2daabef54cc5', 'マジか！食事で何に気をつければいい？', NULL),
	('a0000015-0001-4000-a000-000000000004', 'fa2a8e34-234b-4d40-acad-e4448d1476d0', 3, '41a07087-e2f8-46a9-a9d4-213d62964b30', '発酵食品と食物繊維を意識するだけで、だいぶ変わりますよ。', NULL);

-- Episode 21: J-POP の歴史を辿る（test_user7 / ミュージックステーション）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000016-0001-4000-a000-000000000001', 'a787628e-8dd9-4cac-853d-8d84b0aec6ec', 0, '41a07087-e2f8-46a9-a9d4-213d62964b30', '今日は J-POP の歴史を 90 年代から振り返っていきましょう！', NULL),
	('a0000016-0001-4000-a000-000000000002', 'a787628e-8dd9-4cac-853d-8d84b0aec6ec', 1, '825f7a68-0c4d-453a-be8e-2daabef54cc5', '90 年代といえば小室ファミリーだよね！', NULL),
	('a0000016-0001-4000-a000-000000000003', 'a787628e-8dd9-4cac-853d-8d84b0aec6ec', 2, '41a07087-e2f8-46a9-a9d4-213d62964b30', 'ミリオンセラーが続出した時代でしたね。今とは音楽の届け方が全然違います。', NULL),
	('a0000016-0001-4000-a000-000000000004', 'a787628e-8dd9-4cac-853d-8d84b0aec6ec', 3, '825f7a68-0c4d-453a-be8e-2daabef54cc5', 'サブスク時代になって、音楽の聴き方も変わったよね。', NULL);

-- Episode 22: 戦国武将の意外な一面（test_user8 / 歴史探訪）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000017-0001-4000-a000-000000000001', '2267a065-9958-493b-9cb3-c62e578b7c23', 0, 'f12ff841-52b7-419d-b18f-34b11f0a9b42', '教科書では厳めしい印象の武将たちですが、意外な一面があるんですよ。', NULL),
	('a0000017-0001-4000-a000-000000000002', '2267a065-9958-493b-9cb3-c62e578b7c23', 1, '30c345e4-490b-4b30-8796-120e7643bb5a', 'え、例えばどんなことですか？', NULL),
	('a0000017-0001-4000-a000-000000000003', '2267a065-9958-493b-9cb3-c62e578b7c23', 2, 'f12ff841-52b7-419d-b18f-34b11f0a9b42', '伊達政宗は料理好きで、自分でレシピを考えていたんだよ。', '楽しそうに'),
	('a0000017-0001-4000-a000-000000000004', '2267a065-9958-493b-9cb3-c62e578b7c23', 3, '30c345e4-490b-4b30-8796-120e7643bb5a', 'えー！戦国武将がお料理！ギャップがすごい！', '驚きながら');

-- Episode 23: 古代エジプトの謎（test_user8 / 歴史探訪）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000018-0001-4000-a000-000000000001', '3c8a7f95-0001-4437-a1ec-138009cd0001', 0, 'f12ff841-52b7-419d-b18f-34b11f0a9b42', '古代エジプトのピラミッド、一体どうやって建てたのか？', NULL),
	('a0000018-0001-4000-a000-000000000002', '3c8a7f95-0001-4437-a1ec-138009cd0001', 1, '30c345e4-490b-4b30-8796-120e7643bb5a', '宇宙人が建てたっていう説もありますよね？', NULL),
	('a0000018-0001-4000-a000-000000000003', '3c8a7f95-0001-4437-a1ec-138009cd0001', 2, 'f12ff841-52b7-419d-b18f-34b11f0a9b42', '実は最新の研究で、水を使ったそりで石を運んだことが分かっているんだよ。', NULL),
	('a0000018-0001-4000-a000-000000000004', '3c8a7f95-0001-4437-a1ec-138009cd0001', 3, '30c345e4-490b-4b30-8796-120e7643bb5a', '科学的に解明されていくのが面白いですね！', NULL);

-- Episode 24: 笑ってはいけない早口言葉（test_user9 / コメディナイト）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000019-0001-4000-a000-000000000001', '3c8a7f95-0005-4437-a1ec-138009cd0005', 0, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'よーし、今日は早口言葉チャレンジやるぞ！', NULL),
	('a0000019-0001-4000-a000-000000000002', '3c8a7f95-0005-4437-a1ec-138009cd0005', 1, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', 'はいはい、またくだらないことを。', NULL),
	('a0000019-0001-4000-a000-000000000003', '3c8a7f95-0005-4437-a1ec-138009cd0005', 2, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', '生麦生米生卵！...あれ、噛んでない？', '笑いながら'),
	('a0000019-0001-4000-a000-000000000004', '3c8a7f95-0005-4437-a1ec-138009cd0005', 3, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', '思いっきり噛んでるやないか！', NULL);

-- Episode 25: あるあるネタ大会（test_user9 / コメディナイト）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000020-0001-4000-a000-000000000001', '3c8a7f95-0006-4437-a1ec-138009cd0006', 0, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', '今日は日常のあるあるネタで盛り上がりましょう。', NULL),
	('a0000020-0001-4000-a000-000000000002', '3c8a7f95-0006-4437-a1ec-138009cd0006', 1, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'エレベーターで「閉」ボタン連打しがち！', NULL),
	('a0000020-0001-4000-a000-000000000003', '3c8a7f95-0006-4437-a1ec-138009cd0006', 2, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', 'あるある！あと、間違えて「開」押しちゃうやつね。', '笑いながら'),
	('a0000020-0001-4000-a000-000000000004', '3c8a7f95-0006-4437-a1ec-138009cd0006', 3, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'それやると気まずい空気流れるよねー！', NULL);

-- Episode 26: 数学を好きになる方法（test_user9 / 教育チャンネル）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000021-0001-4000-a000-000000000001', '3c8a7f95-0007-4437-a1ec-138009cd0007', 0, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', '数学が苦手な人って多いけど、実は考え方次第で好きになれるんです。', NULL),
	('a0000021-0001-4000-a000-000000000002', '3c8a7f95-0007-4437-a1ec-138009cd0007', 1, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'えー、ほんとに？ワイ、数学大の苦手やったで。', NULL),
	('a0000021-0001-4000-a000-000000000003', '3c8a7f95-0007-4437-a1ec-138009cd0007', 2, 'b6148fa4-6e9e-4025-a791-b65a327b58e3', '日常の中に数学を見つけるのがコツ。例えばお買い物の割引計算とか。', NULL),
	('a0000021-0001-4000-a000-000000000004', '3c8a7f95-0007-4437-a1ec-138009cd0007', 3, 'ef1c418c-8c10-4b02-887b-2ab1ae937f56', 'あ、割引の計算は得意！それも数学なんか！', '驚きながら');

-- Episode 27: 星降る夜の物語（test_user10 / フィクション工房）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000022-0001-4000-a000-000000000001', '3c8a7f95-0008-4437-a1ec-138009cd0008', 0, '13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2', 'むかしむかし、星が降る村がありました。', NULL),
	('a0000022-0001-4000-a000-000000000002', '3c8a7f95-0008-4437-a1ec-138009cd0008', 1, '5b02ebec-6b83-431a-ae0e-29611d0c7f2f', '少女は毎晩、屋根の上で星を集めていたのです。', NULL),
	('a0000022-0001-4000-a000-000000000003', '3c8a7f95-0008-4437-a1ec-138009cd0008', 2, '13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2', '集めた星を瓶に入れると、淡い光を放ちました。', NULL),
	('a0000022-0001-4000-a000-000000000004', '3c8a7f95-0008-4437-a1ec-138009cd0008', 3, '5b02ebec-6b83-431a-ae0e-29611d0c7f2f', 'その光は、村の人々の願いを叶える不思議な力を持っていたのです。', NULL);

-- Episode 28: 猫と魔法使い（test_user10 / フィクション工房）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0000023-0001-4000-a000-000000000001', '3c8a7f95-0009-4437-a1ec-138009cd0009', 0, '5b02ebec-6b83-431a-ae0e-29611d0c7f2f', '黒猫のクロは、実は魔法使いでした。', NULL),
	('a0000023-0001-4000-a000-000000000002', '3c8a7f95-0009-4437-a1ec-138009cd0009', 1, '13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2', 'ある日、クロは人間の少年と出会います。', NULL),
	('a0000023-0001-4000-a000-000000000003', '3c8a7f95-0009-4437-a1ec-138009cd0009', 2, '5b02ebec-6b83-431a-ae0e-29611d0c7f2f', '「ニャーオ」クロが鳴くと、空から花びらが降ってきました。', NULL),
	('a0000023-0001-4000-a000-000000000004', '3c8a7f95-0009-4437-a1ec-138009cd0009', 3, '13efec0a-fcd2-4a02-a4bd-9e9b7b6afcf2', '少年は目を丸くしました。「きみ、魔法使いなの？」', NULL);

-- Episode 29: プログラミング言語の選び方（test_user / テックトーク）
INSERT INTO script_lines (id, episode_id, line_order, speaker_id, text, emotion) VALUES
	('a0100001-1001-4000-a000-000000000001', 'a0100001-0001-4000-a000-000000000001', 0, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'こんにちは、テックトークへようこそ！今日のテーマは「プログラミング言語の選び方」です。', NULL),
	('a0100001-1001-4000-a000-000000000002', 'a0100001-0001-4000-a000-000000000001', 1, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'よろしくお願いします！実は私、最近プログラミングを始めたいなって思ってるんですけど、言語が多すぎて何を選べばいいか分からなくて。', NULL),
	('a0100001-1001-4000-a000-000000000003', 'a0100001-0001-4000-a000-000000000001', 2, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'あー、それはプログラミング初心者あるあるですね。実は僕も最初めちゃくちゃ悩みました。', '笑いながら'),
	('a0100001-1001-4000-a000-000000000004', 'a0100001-0001-4000-a000-000000000001', 3, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'えっ、ユウキさんでも悩んだんですか？意外です！', '驚いて'),
	('a0100001-1001-4000-a000-000000000005', 'a0100001-0001-4000-a000-000000000001', 4, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'もちろん。まず大事なのは「何を作りたいか」なんですよ。目的によっておすすめの言語が全然変わります。', '真剣に'),
	('a0100001-1001-4000-a000-000000000006', 'a0100001-0001-4000-a000-000000000001', 5, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'なるほど。例えば Web サイトを作りたい場合はどうなりますか？', NULL),
	('a0100001-1001-4000-a000-000000000007', 'a0100001-0001-4000-a000-000000000001', 6, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'Web なら、まず JavaScript は避けて通れないですね。ブラウザで動く唯一のプログラミング言語ですから。', NULL),
	('a0100001-1001-4000-a000-000000000008', 'a0100001-0001-4000-a000-000000000001', 7, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'JavaScript ってよく聞きますよね。でも最近は TypeScript っていうのも人気だって聞いたんですけど。', NULL),
	('a0100001-1001-4000-a000-000000000009', 'a0100001-0001-4000-a000-000000000001', 8, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'いいところに気づきましたね！TypeScript は JavaScript に型という安全装置を付けたものなんです。大きなプロジェクトになると本当に助かります。', NULL),
	('a0100001-1001-4000-a000-00000000000a', 'a0100001-0001-4000-a000-000000000001', 9, '4cee85f3-adec-4333-84e6-d6aaefb63408', '安全装置かぁ。車のシートベルトみたいなものですか？', NULL),
	('a0100001-1001-4000-a000-00000000000b', 'a0100001-0001-4000-a000-000000000001', 10, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'まさにその通り！なくても走れるけど、あった方が安心っていう感じですね。', '笑いながら'),
	('a0100001-1001-4000-a000-00000000000c', 'a0100001-0001-4000-a000-000000000001', 11, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'じゃあ、スマホアプリを作りたい場合はどうですか？', NULL),
	('a0100001-1001-4000-a000-00000000000d', 'a0100001-0001-4000-a000-000000000001', 12, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'iPhone なら Swift、Android なら Kotlin がそれぞれの公式言語です。両方同時に作りたいなら、Flutter の Dart や React Native の JavaScript もありますよ。', '早口で'),
	('a0100001-1001-4000-a000-00000000000e', 'a0100001-0001-4000-a000-000000000001', 13, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'うわ、選択肢が多い！AI とか機械学習に興味がある場合は？', '驚いて'),
	('a0100001-1001-4000-a000-00000000000f', 'a0100001-0001-4000-a000-000000000001', 14, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'AI なら Python 一択と言ってもいいくらいです。ライブラリが圧倒的に充実していて、書き方もシンプルなので初心者にもおすすめです。', '真剣に'),
	('a0100001-1001-4000-a000-000000000010', 'a0100001-0001-4000-a000-000000000001', 15, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'Python！名前は聞いたことあります。蛇のマークのやつですよね。', NULL),
	('a0100001-1001-4000-a000-000000000011', 'a0100001-0001-4000-a000-000000000001', 16, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'そうそう。ちなみに名前の由来は蛇じゃなくて、イギリスのコメディ番組「モンティ・パイソン」なんですよ。', 'ささやいて'),
	('a0100001-1001-4000-a000-000000000012', 'a0100001-0001-4000-a000-000000000001', 17, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'えっ、そうなんですか！面白い！じゃあユウキさんが今一番好きな言語は何ですか？', '嬉しそうに'),
	('a0100001-1001-4000-a000-000000000013', 'a0100001-0001-4000-a000-000000000001', 18, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '僕は最近 Go にハマっています。シンプルで速くて、サーバーサイドの開発にすごく向いているんです。', '嬉しそうに'),
	('a0100001-1001-4000-a000-000000000014', 'a0100001-0001-4000-a000-000000000001', 19, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'Go って Google が作った言語ですよね？', NULL),
	('a0100001-1001-4000-a000-000000000015', 'a0100001-0001-4000-a000-000000000001', 20, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'そうです。コンパイルが速いし、並行処理も簡単に書ける。マスコットのゴーファーも可愛いんですよ。', '嬉しそうに'),
	('a0100001-1001-4000-a000-000000000016', 'a0100001-0001-4000-a000-000000000001', 21, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'マスコットで言語を選ぶのもありですか？', '笑いながら'),
	('a0100001-1001-4000-a000-000000000017', 'a0100001-0001-4000-a000-000000000001', 22, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'モチベーション維持には意外と大事ですよ！冗談抜きで。', '笑いながら'),
	('a0100001-1001-4000-a000-000000000018', 'a0100001-0001-4000-a000-000000000001', 23, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'ところで、最近話題の Rust っていう言語はどうなんですか？難しいって聞くんですけど。', NULL),
	('a0100001-1001-4000-a000-000000000019', 'a0100001-0001-4000-a000-000000000001', 24, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'Rust は確かに学習コストは高いんですが、メモリ安全性が素晴らしくて、バグが起きにくいコードが書けます。システムプログラミングをやりたい人には最高の選択肢です。', '真剣に'),
	('a0100001-1001-4000-a000-00000000001a', 'a0100001-0001-4000-a000-000000000001', 25, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'うーん、色々あって迷いますね。結局、初心者はどの言語から始めるのがベストなんでしょう？', '考えながら'),
	('a0100001-1001-4000-a000-00000000001b', 'a0100001-0001-4000-a000-000000000001', 26, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '個人的には Python か JavaScript をおすすめします。どちらも情報が豊富で、学びやすくて、実用的なものがすぐに作れます。', NULL),
	('a0100001-1001-4000-a000-00000000001c', 'a0100001-0001-4000-a000-000000000001', 27, '4cee85f3-adec-4333-84e6-d6aaefb63408', '私は Web サイトも作りたいし AI にも興味があるので、両方やっちゃおうかな！', '大声で'),
	('a0100001-1001-4000-a000-00000000001d', 'a0100001-0001-4000-a000-000000000001', 28, 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'その意気です！大事なのは、とにかく手を動かして何か作ってみること。完璧な言語選びよりも、まず始めることが一番大切です。', NULL),
	('a0100001-1001-4000-a000-00000000001e', 'a0100001-0001-4000-a000-000000000001', 29, '4cee85f3-adec-4333-84e6-d6aaefb63408', 'はい！今日からさっそく始めてみます。リスナーの皆さんも、ぜひ自分に合った言語を見つけてみてくださいね！それでは、また次回お会いしましょう！', '嬉しそうに');

-- ===========================================
-- 高評価・ブックマーク・フォロー
-- ===========================================

-- test_user (8def69af) が 5 人のユーザーをフォロー
-- フォロー先: test_user2, test_user3, test_user4, test_user5, test_user6
INSERT INTO follows (user_id, target_user_id) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596'),
	('8def69af-dae9-4641-a0e5-100107626933', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5'),
	('8def69af-dae9-4641-a0e5-100107626933', 'd6f829bf-e9bd-4df7-a9f6-64689fa6fcc1'),
	('8def69af-dae9-4641-a0e5-100107626933', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0'),
	('8def69af-dae9-4641-a0e5-100107626933', '80adf759-b01c-4726-87b4-7b9c659483a4');

-- 4 人が test_user をフォロー
-- フォロワー: test_user2, test_user3, test_user7, test_user9
INSERT INTO follows (user_id, target_user_id) VALUES
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', '8def69af-dae9-4641-a0e5-100107626933'),
	('4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', '8def69af-dae9-4641-a0e5-100107626933'),
	('8450d256-8630-4044-8a69-fc8671e6e5c1', '8def69af-dae9-4641-a0e5-100107626933'),
	('c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', '8def69af-dae9-4641-a0e5-100107626933');

-- 全ユーザーにデフォルト再生リスト「後で聴く」を作成
INSERT INTO playlists (id, user_id, name, is_default) VALUES
	('11111111-1111-1111-1111-111111111111', '8def69af-dae9-4641-a0e5-100107626933', '後で聴く', true),
	('22222222-2222-2222-2222-222222222222', '8eada3a5-f413-4eeb-9cd5-12def60d4596', '後で聴く', true),
	('33333333-3333-3333-3333-333333333333', '4dbc55c2-1d78-4e75-b6ac-b5e2b0d461f5', '後で聴く', true),
	('44444444-4444-4444-4444-444444444444', 'd6f829bf-e9bd-4df7-a9f6-64689fa6fcc1', '後で聴く', true),
	('55555555-5555-5555-5555-555555555555', 'b8ad04fd-9afa-474a-a567-1f19e8bcf6b0', '後で聴く', true),
	('66666666-6666-6666-6666-666666666666', '80adf759-b01c-4726-87b4-7b9c659483a4', '後で聴く', true),
	('77777777-7777-7777-7777-777777777777', '8450d256-8630-4044-8a69-fc8671e6e5c1', '後で聴く', true),
	('88888888-8888-8888-8888-888888888888', '7d4e20d4-98ca-4b79-901d-c51e43c38e2f', '後で聴く', true),
	('99999999-9999-9999-9999-999999999999', 'c878a2b4-ade5-44d3-b8ec-d5be985f6dcb', '後で聴く', true),
	('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '767b5ed0-a663-437a-9cc9-b8cef6d0731e', '後で聴く', true);

-- test_user が 5 個のエピソードを「後で聴く」に追加
INSERT INTO playlist_items (playlist_id, episode_id, position) VALUES
	('11111111-1111-1111-1111-111111111111', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 0),
	('11111111-1111-1111-1111-111111111111', '682c05ed-3ae8-4558-a74c-62f8cdc7b344', 1),
	('11111111-1111-1111-1111-111111111111', 'c4d39d21-4e39-4150-bd3d-5ceb27718a38', 2),
	('11111111-1111-1111-1111-111111111111', '105d9802-0133-4a06-a1b4-de87eb26f5ac', 3),
	('11111111-1111-1111-1111-111111111111', '2267a065-9958-493b-9cb3-c62e578b7c23', 4);

-- test_user が 5 個のエピソードに高評価（reaction: like）
INSERT INTO reactions (user_id, episode_id, reaction_type) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 'like'),
	('8def69af-dae9-4641-a0e5-100107626933', 'b1c1e7d7-b3eb-4783-82d0-6857832daf09', 'like'),
	('8def69af-dae9-4641-a0e5-100107626933', '436043d0-9a74-4541-9639-6b9566125bd7', 'like'),
	('8def69af-dae9-4641-a0e5-100107626933', 'b7a74a63-a9e2-4f60-ba58-bfa887598e07', 'like'),
	('8def69af-dae9-4641-a0e5-100107626933', '2267a065-9958-493b-9cb3-c62e578b7c23', 'like');

-- ===========================================
-- コメント
-- ===========================================

-- test_user が他のユーザーのエピソードにコメント
INSERT INTO comments (user_id, episode_id, content) VALUES
	('8def69af-dae9-4641-a0e5-100107626933', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 'とても参考になりました！副業から始めるという考え方が新鮮でした。'),
	('8def69af-dae9-4641-a0e5-100107626933', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', '資金調達の話、勉強になります。');

-- test_user2 が test_user のエピソードにコメント
INSERT INTO comments (user_id, episode_id, content) VALUES
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', 'eb960304-f86e-4364-be5d-d3d5126c9601', 'AIの未来について、とても興味深い内容でした！'),
	('8eada3a5-f413-4eeb-9cd5-12def60d4596', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 'スマートホーム、私も導入を検討してみます。');

-- ===========================================
-- チャンネル公開設定（約 80% = 12/15 チャンネルを公開）
-- 未公開: ゆるふわ雑談ラジオ, ミュージックステーション, フィクション工房
-- ===========================================

UPDATE channels SET published_at = NOW() - INTERVAL '30 days' WHERE id = 'ea9a266e-f532-417c-8916-709d0233941c'; -- テックトーク
UPDATE channels SET published_at = NOW() - INTERVAL '28 days' WHERE id = 'e5a50bd3-8990-4344-b470-56fa7329d75c'; -- ビジネス最前線
UPDATE channels SET published_at = NOW() - INTERVAL '25 days' WHERE id = '3908d99b-d52d-4a73-96fb-fca7df2dfec9'; -- サイエンス・ラボ
UPDATE channels SET published_at = NOW() - INTERVAL '24 days' WHERE id = '395f3bfa-031e-4d90-a53b-19d311392b00'; -- ほのぼのライフ
UPDATE channels SET published_at = NOW() - INTERVAL '22 days' WHERE id = '7f2c8688-c163-40c3-9303-56b4d8a1ed7a'; -- ニュースの裏側
UPDATE channels SET published_at = NOW() - INTERVAL '20 days' WHERE id = '7af3be1e-1555-4f71-9af0-2be26cd6b612'; -- 映画レビュー倶楽部
UPDATE channels SET published_at = NOW() - INTERVAL '18 days' WHERE id = '892d53ff-0633-4b85-a220-afa0682ee467'; -- アート散歩
UPDATE channels SET published_at = NOW() - INTERVAL '15 days' WHERE id = 'b77286af-042a-4580-88b6-efe62aaa3eae'; -- スポーツダイジェスト
UPDATE channels SET published_at = NOW() - INTERVAL '12 days' WHERE id = 'f602c66d-6111-495a-8b9c-e46fa3740d49'; -- ヘルシーライフ
UPDATE channels SET published_at = NOW() - INTERVAL '10 days' WHERE id = 'f02d6a25-4166-4632-812d-446ee6d5be38'; -- 歴史探訪
UPDATE channels SET published_at = NOW() - INTERVAL '7 days'  WHERE id = '1737fa8f-75ff-41ba-9775-8d3496d38344'; -- コメディナイト
UPDATE channels SET published_at = NOW() - INTERVAL '5 days'  WHERE id = '06891da2-4124-492f-826c-1072dcf27ee6'; -- 教育チャンネル

-- ===========================================
-- エピソード用音声データ
-- ===========================================

-- full_audio（BGM ミキシング済み音声）
INSERT INTO audios (id, mime_type, path, filename, file_size, duration_ms) VALUES
	('b0a00001-0001-4000-b000-000000000001', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000002', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000003', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000004', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000005', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000006', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000007', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000008', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000009', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000a', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000b', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000c', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000d', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000e', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000000f', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000010', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000011', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000012', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000013', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000014', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000015', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000016', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000017', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000018', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000019', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001a', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001b', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001c', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001d', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001e', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-00000000001f', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000020', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000021', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000022', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000023', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000024', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00001-0001-4000-b000-000000000025', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000);

-- voice_audio（TTS のみの音声）
INSERT INTO audios (id, mime_type, path, filename, file_size, duration_ms) VALUES
	('b0a00002-0001-4000-b000-000000000001', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000002', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000003', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000004', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000005', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000006', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000007', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000008', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000009', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000a', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000b', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000c', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000d', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000e', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000000f', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000010', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000011', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000012', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000013', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000014', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000015', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000016', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000017', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000018', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000019', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001a', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001b', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001c', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001d', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001e', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-00000000001f', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000020', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000021', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000022', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000023', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000024', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000),
	('b0a00002-0001-4000-b000-000000000025', 'audio/mpeg', 'system/sample_audio.mp3', 'sample_audio.mp3', 153600, 8000);

-- エピソードに音声を紐づけ
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000001', voice_audio_id = 'b0a00002-0001-4000-b000-000000000001' WHERE id = 'eb960304-f86e-4364-be5d-d3d5126c9601'; -- AI の未来を語る
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000002', voice_audio_id = 'b0a00002-0001-4000-b000-000000000002' WHERE id = '67e8e26d-20c8-492a-ac2c-5c79d8050aa3'; -- スマートホームのすすめ
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000003', voice_audio_id = 'b0a00002-0001-4000-b000-000000000003' WHERE id = '198d7e19-7d40-4299-95bf-a641f5c83911'; -- 最近ハマってること
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000004', voice_audio_id = 'b0a00002-0001-4000-b000-000000000004' WHERE id = 'fcb16526-951a-4ff1-a456-ab1dba96f699'; -- 副業から始める起業入門
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000005', voice_audio_id = 'b0a00002-0001-4000-b000-000000000005' WHERE id = '9cde2abb-30e8-447b-bc8b-bb799b0f6f06'; -- 失敗しない資金調達の秘訣
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000006', voice_audio_id = 'b0a00002-0001-4000-b000-000000000006' WHERE id = 'b1c1e7d7-b3eb-4783-82d0-6857832daf09'; -- 量子コンピュータ入門
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000007', voice_audio_id = 'b0a00002-0001-4000-b000-000000000007' WHERE id = '682c05ed-3ae8-4558-a74c-62f8cdc7b344'; -- 宇宙の神秘に迫る
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000008', voice_audio_id = 'b0a00002-0001-4000-b000-000000000008' WHERE id = '1a4aad00-fd65-4960-a11f-45f2b5ff504c'; -- DNA と遺伝子の不思議
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000009', voice_audio_id = 'b0a00002-0001-4000-b000-000000000009' WHERE id = '98f515d1-ca1f-4810-b8b6-a147aac641f4'; -- お気に入りのカフェ巡り
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000a', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000a' WHERE id = '29c31347-3078-4ab0-9773-f6d408462c39'; -- 休日の過ごし方
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000b', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000b' WHERE id = '436043d0-9a74-4541-9639-6b9566125bd7'; -- SNS 時代のメディアリテラシー
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000c', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000c' WHERE id = 'c4d39d21-4e39-4150-bd3d-5ceb27718a38'; -- 選挙と民主主義の未来
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000d', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000d' WHERE id = '8cfcfff0-87c8-4db9-a250-7ccc5fba1407'; -- 気候変動と私たちの暮らし
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000e', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000e' WHERE id = '7c2d9c7a-e3e3-4388-bfe7-d36162a19768'; -- リモートワーク革命
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000000f', voice_audio_id = 'b0a00002-0001-4000-b000-00000000000f' WHERE id = 'b7a74a63-a9e2-4f60-ba58-bfa887598e07'; -- 今年のベスト映画 TOP5
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000010', voice_audio_id = 'b0a00002-0001-4000-b000-000000000010' WHERE id = '105d9802-0133-4a06-a1b4-de87eb26f5ac'; -- ホラー映画の魅力
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000011', voice_audio_id = 'b0a00002-0001-4000-b000-000000000011' WHERE id = 'ec47dd8e-34c5-4ba4-a101-02713e87ba74'; -- 現代アートの楽しみ方
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000012', voice_audio_id = 'b0a00002-0001-4000-b000-000000000012' WHERE id = '4697c3c7-a89a-4ce5-9b39-f4bb8d6ed69a'; -- サッカー W 杯を振り返る
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000013', voice_audio_id = 'b0a00002-0001-4000-b000-000000000013' WHERE id = 'ecd6cb92-597d-4367-9c65-9f98fb38fc9b'; -- マラソンの科学
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000014', voice_audio_id = 'b0a00002-0001-4000-b000-000000000014' WHERE id = 'f270e88b-4c7f-4fd5-a06d-5a52ab0ff501'; -- 筋トレとメンタルの関係
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000015', voice_audio_id = 'b0a00002-0001-4000-b000-000000000015' WHERE id = '80d5ac1d-4a0c-44b9-8e4c-61b2ac333b00'; -- eスポーツの今
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000016', voice_audio_id = 'b0a00002-0001-4000-b000-000000000016' WHERE id = '948ee16d-ef1b-4e54-8dcb-cf6b7b46c7fd'; -- 野球データ分析入門
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000017', voice_audio_id = 'b0a00002-0001-4000-b000-000000000017' WHERE id = '4288967f-f224-452a-bf4d-f0c61a1ecaa9'; -- オリンピックの感動秘話
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000018', voice_audio_id = 'b0a00002-0001-4000-b000-000000000018' WHERE id = '9890df9b-05cd-4962-9c7f-42e05b5c6ced'; -- 朝ヨガのすすめ
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000019', voice_audio_id = 'b0a00002-0001-4000-b000-000000000019' WHERE id = 'fa2a8e34-234b-4d40-acad-e4448d1476d0'; -- 腸活で健康に
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001a', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001a' WHERE id = 'a787628e-8dd9-4cac-853d-8d84b0aec6ec'; -- J-POP の歴史を辿る
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001b', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001b' WHERE id = '2267a065-9958-493b-9cb3-c62e578b7c23'; -- 戦国武将の意外な一面
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001c', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001c' WHERE id = '3c8a7f95-0001-4437-a1ec-138009cd0001'; -- 古代エジプトの謎
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001d', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001d' WHERE id = '3c8a7f95-0002-4437-a1ec-138009cd0002'; -- 幕末の志士たち
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001e', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001e' WHERE id = '3c8a7f95-0003-4437-a1ec-138009cd0003'; -- ローマ帝国の栄光と衰退
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-00000000001f', voice_audio_id = 'b0a00002-0001-4000-b000-00000000001f' WHERE id = '3c8a7f95-0004-4437-a1ec-138009cd0004'; -- 日本の城の秘密
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000020', voice_audio_id = 'b0a00002-0001-4000-b000-000000000020' WHERE id = '3c8a7f95-0005-4437-a1ec-138009cd0005'; -- 笑ってはいけない早口言葉
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000021', voice_audio_id = 'b0a00002-0001-4000-b000-000000000021' WHERE id = '3c8a7f95-0006-4437-a1ec-138009cd0006'; -- あるあるネタ大会
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000022', voice_audio_id = 'b0a00002-0001-4000-b000-000000000022' WHERE id = '3c8a7f95-0007-4437-a1ec-138009cd0007'; -- 数学を好きになる方法
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000023', voice_audio_id = 'b0a00002-0001-4000-b000-000000000023' WHERE id = '3c8a7f95-0008-4437-a1ec-138009cd0008'; -- 星降る夜の物語
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000024', voice_audio_id = 'b0a00002-0001-4000-b000-000000000024' WHERE id = '3c8a7f95-0009-4437-a1ec-138009cd0009'; -- 猫と魔法使い
UPDATE episodes SET full_audio_id = 'b0a00001-0001-4000-b000-000000000025', voice_audio_id = 'b0a00002-0001-4000-b000-000000000025' WHERE id = '50100001-0003-4000-a000-000000000001'; -- 朝の散歩で見つけたこと
