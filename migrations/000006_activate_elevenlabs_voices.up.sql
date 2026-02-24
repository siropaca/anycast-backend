-- 既存 ElevenLabs ボイスを有効化
UPDATE voices SET is_active = true WHERE provider = 'elevenlabs' AND provider_voice_id IN (
	'b34JylakFZPlGS0BnwyY',
	'4lOQ7A2l7HPuG7UIHiKA',
	'JTlYtJrcTzPC71hMLOxo',
	'NO5A3b3sSzDyJQF7MiNS'
);
