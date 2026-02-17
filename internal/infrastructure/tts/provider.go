package tts

// Provider は TTS プロバイダの種別
type Provider string

const (
	ProviderGoogle     Provider = "google"
	ProviderElevenLabs Provider = "elevenlabs"
)
