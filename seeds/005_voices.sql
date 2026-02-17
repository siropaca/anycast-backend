-- 開発環境用: Google ボイスを有効化
UPDATE voices SET is_active = true WHERE provider = 'google';
