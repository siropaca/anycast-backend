-- 既存 ElevenLabs ボイスを無効化に戻す
UPDATE voices SET is_active = false WHERE provider = 'elevenlabs' AND provider_voice_id IN (
	'b34JylakFZPlGS0BnwyY',
	'4lOQ7A2l7HPuG7UIHiKA',
	'JTlYtJrcTzPC71hMLOxo',
	'NO5A3b3sSzDyJQF7MiNS'
);
