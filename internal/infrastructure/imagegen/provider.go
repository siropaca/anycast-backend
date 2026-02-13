package imagegen

// Provider は画像生成プロバイダの種別
type Provider string

const (
	ProviderGemini Provider = "gemini"
	ProviderOpenAI Provider = "openai"
)
