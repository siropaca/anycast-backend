-- システム BGM の seed データを削除（system_bgms → audios の順で削除）
DELETE FROM system_bgms;
DELETE FROM audios WHERE path LIKE 'audios/%' AND path IN (
	'audios/You_and_Me.mp3',
	'audios/2_23_AM.mp3',
	'audios/しゅわしゅわハニーレモン350ml.mp3',
	'audios/10℃.mp3',
	'audios/SUMMER_TRIANGLE.mp3',
	'audios/pastel_house.mp3',
	'audios/野良猫は宇宙を目指した.mp3',
	'audios/feeling_like_afternoon.mp3'
);
