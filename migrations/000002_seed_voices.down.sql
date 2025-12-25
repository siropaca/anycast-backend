-- Remove seeded Google TTS voices
DELETE FROM voices WHERE provider = 'google';
