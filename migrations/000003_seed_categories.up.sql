-- カテゴリの seed データ

-- カテゴリ用の画像データ（Unsplash）
INSERT INTO images (id, mime_type, path, filename, file_size) VALUES
	('c0c0c0c0-0001-4000-8000-000000000001', 'image/jpeg', 'https://images.unsplash.com/photo-1618424181497-157f25b6ddd5?w=800&h=800&fit=crop', 'category-technology.jpg', 0),
	('c0c0c0c0-0002-4000-8000-000000000002', 'image/jpeg', 'https://images.unsplash.com/photo-1490367532201-b9bc1dc483f6?w=800&h=800&fit=crop', 'category-business.jpg', 0),
	('c0c0c0c0-0003-4000-8000-000000000003', 'image/jpeg', 'https://images.unsplash.com/photo-1691995539409-a5438d630d1c?w=800&h=800&fit=crop', 'category-news.jpg', 0),
	('c0c0c0c0-0004-4000-8000-000000000004', 'image/jpeg', 'https://images.unsplash.com/photo-1602145095452-aba06946ed05?w=800&h=800&fit=crop', 'category-education.jpg', 0),
	('c0c0c0c0-0005-4000-8000-000000000005', 'image/jpeg', 'https://images.unsplash.com/photo-1607805074620-5802aee47bdb?w=800&h=800&fit=crop', 'category-comedy.jpg', 0),
	('c0c0c0c0-0006-4000-8000-000000000006', 'image/jpeg', 'https://images.unsplash.com/photo-1547142115-5e49d39a40bc?w=800&h=800&fit=crop', 'category-society-culture.jpg', 0),
	('c0c0c0c0-0007-4000-8000-000000000007', 'image/jpeg', 'https://images.unsplash.com/photo-1579783928621-7a13d66a62d1?w=800&h=800&fit=crop', 'category-arts.jpg', 0),
	('c0c0c0c0-0008-4000-8000-000000000008', 'image/jpeg', 'https://images.unsplash.com/photo-1602052577122-f73b9710adba?w=800&h=800&fit=crop', 'category-science.jpg', 0),
	('c0c0c0c0-0009-4000-8000-000000000009', 'image/jpeg', 'https://images.unsplash.com/photo-1634144646738-809a0f8897c4?w=800&h=800&fit=crop', 'category-health-fitness.jpg', 0),
	('c0c0c0c0-0010-4000-8000-000000000010', 'image/jpeg', 'https://images.unsplash.com/photo-1461896836934-ffe607ba8211?w=800&h=800&fit=crop', 'category-sports.jpg', 0),
	('c0c0c0c0-0011-4000-8000-000000000011', 'image/jpeg', 'https://images.unsplash.com/photo-1470225620780-dba8ba36b745?w=800&h=800&fit=crop', 'category-music.jpg', 0),
	('c0c0c0c0-0012-4000-8000-000000000012', 'image/jpeg', 'https://images.unsplash.com/photo-1682130301125-5b63bbf93241?w=800&h=800&fit=crop', 'category-tv-film.jpg', 0),
	('c0c0c0c0-0013-4000-8000-000000000013', 'image/jpeg', 'https://images.unsplash.com/photo-1550399105-c4db5fb85c18?w=800&h=800&fit=crop', 'category-history.jpg', 0),
	('c0c0c0c0-0014-4000-8000-000000000014', 'image/jpeg', 'https://images.unsplash.com/photo-1613399421098-f943ea81f1c4?w=800&h=800&fit=crop', 'category-documentary.jpg', 0),
	('c0c0c0c0-0015-4000-8000-000000000015', 'image/jpeg', 'https://images.unsplash.com/photo-1760448847959-bd3aec9e672c?w=800&h=800&fit=crop', 'category-fiction.jpg', 0),
	('c0c0c0c0-0016-4000-8000-000000000016', 'image/jpeg', 'https://images.unsplash.com/photo-1647616927583-1d44a79a38a5?w=800&h=800&fit=crop', 'category-kids-family.jpg', 0),
	('c0c0c0c0-0017-4000-8000-000000000017', 'image/jpeg', 'https://images.unsplash.com/photo-1631722235421-2f4d55468fdf?w=800&h=800&fit=crop', 'category-leisure.jpg', 0),
	('c0c0c0c0-0018-4000-8000-000000000018', 'image/jpeg', 'https://images.unsplash.com/photo-1607545658310-8fb58058736f?w=800&h=800&fit=crop', 'category-religion-spirituality.jpg', 0),
	('c0c0c0c0-0019-4000-8000-000000000019', 'image/jpeg', 'https://images.unsplash.com/photo-1666798044958-9df7c6cc279a?w=800&h=800&fit=crop', 'category-government.jpg', 0);

INSERT INTO categories (slug, name, image_id, sort_order) VALUES
	('technology', 'テクノロジー', 'c0c0c0c0-0001-4000-8000-000000000001', 1),
	('business', 'ビジネス', 'c0c0c0c0-0002-4000-8000-000000000002', 2),
	('news', 'ニュース', 'c0c0c0c0-0003-4000-8000-000000000003', 3),
	('education', '教育', 'c0c0c0c0-0004-4000-8000-000000000004', 4),
	('comedy', 'コメディ', 'c0c0c0c0-0005-4000-8000-000000000005', 5),
	('society-culture', '社会/文化', 'c0c0c0c0-0006-4000-8000-000000000006', 6),
	('arts', 'アート', 'c0c0c0c0-0007-4000-8000-000000000007', 7),
	('science', 'サイエンス', 'c0c0c0c0-0008-4000-8000-000000000008', 8),
	('health-fitness', '健康/フィットネス', 'c0c0c0c0-0009-4000-8000-000000000009', 9),
	('sports', 'スポーツ', 'c0c0c0c0-0010-4000-8000-000000000010', 10),
	('music', 'ミュージック', 'c0c0c0c0-0011-4000-8000-000000000011', 11),
	('tv-film', 'TV/映画', 'c0c0c0c0-0012-4000-8000-000000000012', 12),
	('history', '歴史', 'c0c0c0c0-0013-4000-8000-000000000013', 13),
	('documentary', 'ドキュメンタリー', 'c0c0c0c0-0014-4000-8000-000000000014', 14),
	('fiction', 'フィクション', 'c0c0c0c0-0015-4000-8000-000000000015', 15),
	('kids-family', 'キッズ/ファミリー', 'c0c0c0c0-0016-4000-8000-000000000016', 16),
	('leisure', 'レジャー', 'c0c0c0c0-0017-4000-8000-000000000017', 17),
	('religion-spirituality', '宗教/スピリチュアル', 'c0c0c0c0-0018-4000-8000-000000000018', 18),
	('government', '行政', 'c0c0c0c0-0019-4000-8000-000000000019', 19);
