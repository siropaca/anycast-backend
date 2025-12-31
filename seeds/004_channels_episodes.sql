-- テスト用のチャンネル・エピソード・台本データを作成する

-- 既存のテストデータを削除（冪等性のため）
-- script_lines.speaker_id は ON DELETE RESTRICT のため、先に削除する必要がある
DELETE FROM script_lines WHERE episode_id IN (
	SELECT e.id FROM episodes e
	JOIN channels c ON e.channel_id = c.id
	WHERE c.user_id IN ('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596')
);
-- likes, bookmarks, follows はテストユーザーのものを削除
DELETE FROM likes WHERE user_id IN ('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596');
DELETE FROM bookmarks WHERE user_id IN ('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596');
DELETE FROM follows WHERE user_id IN ('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596');
DELETE FROM channels WHERE user_id IN ('8def69af-dae9-4641-a0e5-100107626933', '8eada3a5-f413-4eeb-9cd5-12def60d4596');

-- ===========================================
-- チャンネル
-- ===========================================

-- test_user のチャンネル
INSERT INTO channels (id, user_id, name, description, category_id, artwork_id) VALUES
	('ea9a266e-f532-417c-8916-709d0233941c', '8def69af-dae9-4641-a0e5-100107626933', 'テックトーク', '最新のテクノロジーニュースを2人のパーソナリティが楽しく解説するポッドキャスト', (SELECT id FROM categories WHERE slug = 'technology'), '4946f33c-3c66-40ca-8b35-3bbdfe65b20c'),
	('efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', '8def69af-dae9-4641-a0e5-100107626933', 'ゆるふわ雑談ラジオ', '日常のあれこれをゆるく語る雑談番組', (SELECT id FROM categories WHERE slug = 'society-culture'), '9ee172c8-6deb-4598-a379-d7fdf502db9a');

-- test_user2 のチャンネル
INSERT INTO channels (id, user_id, name, description, category_id, artwork_id) VALUES
	('e5a50bd3-8990-4344-b470-56fa7329d75c', '8eada3a5-f413-4eeb-9cd5-12def60d4596', 'ビジネス最前線', '起業やキャリアについて実践的なアドバイスを届けるビジネス番組', (SELECT id FROM categories WHERE slug = 'business'), NULL);

-- ===========================================
-- キャラクター
-- ===========================================

-- テックトークのキャラクター
INSERT INTO characters (id, channel_id, name, persona, voice_id) VALUES
	('d1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'ea9a266e-f532-417c-8916-709d0233941c', 'ユウキ', 'テクノロジーに詳しいエンジニア。論理的だが親しみやすい話し方をする。', (SELECT id FROM voices WHERE name = 'Achird')),
	('4cee85f3-adec-4333-84e6-d6aaefb63408', 'ea9a266e-f532-417c-8916-709d0233941c', 'ミサキ', '好奇心旺盛なライター。素朴な疑問を投げかけてくれる。', (SELECT id FROM voices WHERE name = 'Achernar'));

-- ゆるふわ雑談ラジオのキャラクター
INSERT INTO characters (id, channel_id, name, persona, voice_id) VALUES
	('b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', 'efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', 'ハルカ', 'のんびり屋でマイペース。ゆるい雰囲気で話を進める。', (SELECT id FROM voices WHERE name = 'Aoede')),
	('41977119-13d8-4d26-bfe4-694eb2cf2167', 'efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', 'ソウタ', 'ツッコミ担当。ハルカのボケに的確に反応する。', (SELECT id FROM voices WHERE name = 'Charon'));

-- ビジネス最前線のキャラクター
INSERT INTO characters (id, channel_id, name, persona, voice_id) VALUES
	('a4e0f973-f91a-4103-b758-fed371622046', 'e5a50bd3-8990-4344-b470-56fa7329d75c', 'ケンジ', '元外資系コンサルタント。論理的で鋭い視点を持つ。', (SELECT id FROM voices WHERE name = 'Fenrir')),
	('b7efbbae-0655-46f1-afb7-a42d2646f0c1', 'e5a50bd3-8990-4344-b470-56fa7329d75c', 'アヤカ', 'スタートアップ経営者。実体験に基づいたアドバイスが得意。', (SELECT id FROM voices WHERE name = 'Kore'));

-- ===========================================
-- エピソード
-- ===========================================

-- test_user のエピソード
INSERT INTO episodes (id, channel_id, title, description, status, full_audio_id) VALUES
	('eb960304-f86e-4364-be5d-d3d5126c9601', 'ea9a266e-f532-417c-8916-709d0233941c', 'AI の未来を語る', 'ChatGPT から始まった AI ブームの今後について', 'published', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 'ea9a266e-f532-417c-8916-709d0233941c', 'スマートホームのすすめ', '自宅を便利にするガジェット紹介', 'published', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('198d7e19-7d40-4299-95bf-a641f5c83911', 'efa53d40-2f7b-4abe-a1b0-ba4f7905dbad', '最近ハマってること', 'お互いの趣味について語り合う回', 'draft', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- test_user2 のエピソード
INSERT INTO episodes (id, channel_id, title, description, status, full_audio_id) VALUES
	('fcb16526-951a-4ff1-a456-ab1dba96f699', 'e5a50bd3-8990-4344-b470-56fa7329d75c', '副業から始める起業入門', 'リスクを抑えながら起業にチャレンジする方法', 'published', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 'e5a50bd3-8990-4344-b470-56fa7329d75c', '失敗しない資金調達の秘訣', 'スタートアップの資金調達で気をつけるべきポイント', 'published', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- ===========================================
-- 台本（ScriptLines）
-- ===========================================

-- Episode 1: AI の未来を語る
INSERT INTO script_lines (id, episode_id, line_order, line_type, speaker_id, text, emotion, audio_id) VALUES
	('236f9071-900a-4b75-aea7-ebb847f5ccad', 'eb960304-f86e-4364-be5d-d3d5126c9601', 0, 'speech', 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'こんにちは、テックトークへようこそ！今日は AI の未来について話していきます。', '明るく', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('bff9d166-1ad5-46fa-96fb-39a27378e99f', 'eb960304-f86e-4364-be5d-d3d5126c9601', 1, 'speech', '4cee85f3-adec-4333-84e6-d6aaefb63408', 'よろしくお願いします！最近 ChatGPT がすごく話題ですよね。', '興味深げに', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('d5422671-73c8-4b28-afe1-5b0c419dcd49', 'eb960304-f86e-4364-be5d-d3d5126c9601', 2, 'speech', 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', 'そうなんです。大規模言語モデルの進化は目覚ましいものがあります。', '解説するように', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('51223f66-3ac5-4685-9609-50d0ccd9b10a', 'eb960304-f86e-4364-be5d-d3d5126c9601', 3, 'speech', '4cee85f3-adec-4333-84e6-d6aaefb63408', 'これからどんな未来が待っているのか、楽しみですね！', 'ワクワクしながら', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- Episode 2: スマートホームのすすめ
INSERT INTO script_lines (id, episode_id, line_order, line_type, speaker_id, text, emotion, audio_id) VALUES
	('d8909d91-da04-4ec6-bec1-356eb9c4e2d9', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 0, 'speech', 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '今日はスマートホームについて紹介していきます。', '落ち着いて', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('f0d41215-6172-4bca-a10e-efaa002a09fc', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 1, 'speech', '4cee85f3-adec-4333-84e6-d6aaefb63408', 'スマートホームって難しそうなイメージがあるんですけど…', '不安げに', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('6723d570-d7c6-4a07-b481-0b609765be86', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 2, 'speech', 'd1f7e3e3-d2e9-4a8f-a155-99b1e3ccf007', '実は意外と簡単に始められるんですよ。スマートスピーカーから始めるのがおすすめです。', '優しく', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('504ebb03-a05c-49ce-9e94-974f9cc80cc0', '67e8e26d-20c8-492a-ac2c-5c79d8050aa3', 3, 'speech', '4cee85f3-adec-4333-84e6-d6aaefb63408', 'なるほど！それなら私でもできそうです。', '安心して', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- Episode 3: 最近ハマってること
INSERT INTO script_lines (id, episode_id, line_order, line_type, speaker_id, text, emotion, audio_id) VALUES
	('23e48682-4949-4751-aef0-b80e369a899b', '198d7e19-7d40-4299-95bf-a641f5c83911', 0, 'speech', 'b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', 'ねえねえ、最近なんかハマってることある？', 'のんびりと', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('8b7f30af-1662-413d-84b6-2b27033435f7', '198d7e19-7d40-4299-95bf-a641f5c83911', 1, 'speech', '41977119-13d8-4d26-bfe4-694eb2cf2167', '最近はコーヒーにハマってるかな。豆から挽いて淹れてるよ。', '楽しそうに', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('fd9ce404-74c8-456e-8f1b-eda25e22ccce', '198d7e19-7d40-4299-95bf-a641f5c83911', 2, 'speech', 'b0b67254-ff3b-4b5e-96fa-073ce5c8a6a4', 'へー、おしゃれだね〜。私は最近観葉植物を育て始めたんだ。', 'ほんわかと', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('18641a58-561f-466a-a670-cf3a569c6669', '198d7e19-7d40-4299-95bf-a641f5c83911', 3, 'speech', '41977119-13d8-4d26-bfe4-694eb2cf2167', '植物いいね！どんな種類を育ててるの？', '興味を持って', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- Episode 4: 副業から始める起業入門（test_user2）
INSERT INTO script_lines (id, episode_id, line_order, line_type, speaker_id, text, emotion, audio_id) VALUES
	('ae5f21f0-a737-47cf-8d00-e9f490bea753', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 0, 'speech', 'a4e0f973-f91a-4103-b758-fed371622046', '今日は副業から起業を始める方法についてお話しします。', '真剣に', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('1a87b77a-2211-4421-9f2c-334ce913e5c3', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 1, 'speech', 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '私も最初は副業からスタートしました。いきなり会社を辞めるのはリスクが高いですからね。', '経験を語るように', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('b8f262e5-f027-484a-a0f5-997e5b9dd569', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 2, 'speech', 'a4e0f973-f91a-4103-b758-fed371622046', 'まずは小さく始めて、収益が安定してから本格的に移行するのがおすすめです。', 'アドバイスするように', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('da61ccde-dfea-4ad8-8a84-f8c4d5a79ac3', 'fcb16526-951a-4ff1-a456-ab1dba96f699', 3, 'speech', 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', 'そうですね。失敗しても学びになりますし、挑戦することが大切です。', '励ますように', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- Episode 5: 失敗しない資金調達の秘訣（test_user2）
INSERT INTO script_lines (id, episode_id, line_order, line_type, speaker_id, text, emotion, audio_id) VALUES
	('089e59a2-e26b-4dcc-aeca-6763a7ab16b9', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 0, 'speech', 'a4e0f973-f91a-4103-b758-fed371622046', '今回は資金調達について詳しくお話ししていきます。', '落ち着いて', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('83fccc88-5647-47df-9757-5ebd19b301c7', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 1, 'speech', 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '資金調達って種類がたくさんありますよね。VC、エンジェル投資家、融資…', '考えながら', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('8f197ebe-677f-4d9d-add7-111af58b6c04', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 2, 'speech', 'a4e0f973-f91a-4103-b758-fed371622046', 'その通りです。事業のフェーズによって最適な調達方法は変わってきます。', '解説するように', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2'),
	('2fecb517-5d29-4f74-b3a7-7a85700e4e22', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06', 3, 'speech', 'b7efbbae-0655-46f1-afb7-a42d2646f0c1', '私の場合は最初にエンジェル投資家から調達しました。その経験も後でお話ししますね。', '振り返りながら', '4b33e6ee-c81b-4795-a843-74fd82fa4fd2');

-- ===========================================
-- お気に入り・ブックマーク・フォロー
-- ===========================================

-- test_user が test_user2 のエピソードをお気に入り
INSERT INTO likes (id, user_id, episode_id) VALUES
	('3c8a7f95-c316-4437-a1ec-138009cd0833', '8def69af-dae9-4641-a0e5-100107626933', 'fcb16526-951a-4ff1-a456-ab1dba96f699'),
	('14b3cfbb-c080-4518-985c-6cd5c226900c', '8def69af-dae9-4641-a0e5-100107626933', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06');

-- test_user が test_user2 のエピソードをブックマーク
INSERT INTO bookmarks (id, user_id, episode_id) VALUES
	('1cf6c66c-2efc-4dd6-8000-d13128bb5384', '8def69af-dae9-4641-a0e5-100107626933', 'fcb16526-951a-4ff1-a456-ab1dba96f699'),
	('52bd26d4-d0aa-4d97-9a85-25a22e3df5d5', '8def69af-dae9-4641-a0e5-100107626933', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06');

-- test_user が test_user2 のエピソードをフォロー
INSERT INTO follows (id, user_id, episode_id) VALUES
	('6869b7ad-1859-4b4f-8898-a3229f7ce27d', '8def69af-dae9-4641-a0e5-100107626933', 'fcb16526-951a-4ff1-a456-ab1dba96f699'),
	('024c2206-a2ed-465d-b468-43a40b891264', '8def69af-dae9-4641-a0e5-100107626933', '9cde2abb-30e8-447b-bc8b-bb799b0f6f06');
