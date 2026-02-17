-- voices の seed データを削除
DELETE FROM voices WHERE provider = 'google';
DELETE FROM voices WHERE provider = 'elevenlabs';
