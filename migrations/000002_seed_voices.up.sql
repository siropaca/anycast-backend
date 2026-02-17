-- Google Gemini TTS ボイスの seed データ

INSERT INTO voices (provider, provider_voice_id, name, gender, sample_audio_url, is_active) VALUES
	('google', 'Achernar', 'Achernar', 'female', '/voice/google/achernar.wav', false),
	('google', 'Achird', 'Achird', 'male', '/voice/google/achird.wav', false),
	('google', 'Algenib', 'Algenib', 'male', '/voice/google/algenib.wav', false),
	('google', 'Algieba', 'Algieba', 'male', '/voice/google/algieba.wav', false),
	('google', 'Alnilam', 'Alnilam', 'male', '/voice/google/alnilam.wav', false),
	('google', 'Aoede', 'Aoede', 'female', '/voice/google/aoede.wav', false),
	('google', 'Autonoe', 'Autonoe', 'female', '/voice/google/autonoe.wav', false),
	('google', 'Callirrhoe', 'Callirrhoe', 'female', '/voice/google/callirrhoe.wav', false),
	('google', 'Charon', 'Charon', 'male', '/voice/google/charon.wav', false),
	('google', 'Despina', 'Despina', 'female', '/voice/google/despina.wav', false),
	('google', 'Enceladus', 'Enceladus', 'male', '/voice/google/enceladus.wav', false),
	('google', 'Erinome', 'Erinome', 'female', '/voice/google/erinome.wav', false),
	('google', 'Fenrir', 'Fenrir', 'male', '/voice/google/fenrir.wav', false),
	('google', 'Gacrux', 'Gacrux', 'female', '/voice/google/gacrux.wav', false),
	('google', 'Iapetus', 'Iapetus', 'male', '/voice/google/iapetus.wav', false),
	('google', 'Kore', 'Kore', 'female', '/voice/google/kore.wav', false),
	('google', 'Laomedeia', 'Laomedeia', 'female', '/voice/google/laomedeia.wav', false),
	('google', 'Leda', 'Leda', 'female', '/voice/google/leda.wav', false),
	('google', 'Orus', 'Orus', 'male', '/voice/google/orus.wav', false),
	('google', 'Puck', 'Puck', 'male', '/voice/google/puck.wav', false),
	('google', 'Pulcherrima', 'Pulcherrima', 'female', '/voice/google/pulcherrima.wav', false),
	('google', 'Rasalgethi', 'Rasalgethi', 'male', '/voice/google/rasalgethi.wav', false),
	('google', 'Sadachbia', 'Sadachbia', 'male', '/voice/google/sadachbia.wav', false),
	('google', 'Sadaltager', 'Sadaltager', 'male', '/voice/google/sadaltager.wav', false),
	('google', 'Schedar', 'Schedar', 'male', '/voice/google/schedar.wav', false),
	('google', 'Sulafat', 'Sulafat', 'female', '/voice/google/sulafat.wav', false),
	('google', 'Umbriel', 'Umbriel', 'male', '/voice/google/umbriel.wav', false),
	('google', 'Vindemiatrix', 'Vindemiatrix', 'female', '/voice/google/vindemiatrix.wav', false),
	('google', 'Zephyr', 'Zephyr', 'female', '/voice/google/zephyr.wav', false),
	('google', 'Zubenelgenubi', 'Zubenelgenubi', 'male', '/voice/google/zubenelgenubi.wav', false);

-- ElevenLabs TTS ボイスの seed データ

INSERT INTO voices (provider, provider_voice_id, name, gender, sample_audio_url) VALUES
	('elevenlabs', 'b34JylakFZPlGS0BnwyY', 'Kenzo', 'male', '/voice/elevenlabs/kenzo.wav'),
	('elevenlabs', '4lOQ7A2l7HPuG7UIHiKA', 'Kyoko', 'female', '/voice/elevenlabs/kyoko.wav'),
	('elevenlabs', 'JTlYtJrcTzPC71hMLOxo', 'Yuki', 'female', '/voice/elevenlabs/yuki.wav'),
	('elevenlabs', 'NO5A3b3sSzDyJQF7MiNS', 'Shohei', 'male', '/voice/elevenlabs/shohei.wav');
