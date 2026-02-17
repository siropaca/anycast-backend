package tts

// Provider は TTS プロバイダの種別
type Provider string

const (
	ProviderGemini     Provider = "gemini"
	ProviderElevenLabs Provider = "ElevenLabs"
)
