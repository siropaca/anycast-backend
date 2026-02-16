-- システム BGM の seed データを削除（system_bgms → audios の順で削除）
DELETE FROM system_bgms;
DELETE FROM audios WHERE path LIKE 'system/%' AND path IN (
	'system/You_and_Me.mp3',
	'system/2_23_AM.mp3',
	'system/しゅわしゅわハニーレモン350ml.mp3',
	'system/10℃.mp3',
	'system/SUMMER_TRIANGLE.mp3',
	'system/pastel_house.mp3',
	'system/野良猫は宇宙を目指した.mp3',
	'system/feeling_like_afternoon.mp3'
);
