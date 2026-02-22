-- 開発環境では ElevenLabs ボイスを有効にする
-- 本番マイグレーションでは is_active = false で投入されている

UPDATE voices SET is_active = true WHERE provider = 'elevenlabs';
