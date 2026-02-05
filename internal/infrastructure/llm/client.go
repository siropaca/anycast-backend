package llm

import (
	"context"
	"fmt"
)

// Provider は LLM プロバイダの種別
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderClaude Provider = "claude"
	ProviderGemini Provider = "gemini"
)

// ClientConfig は LLM クライアントの設定
type ClientConfig struct {
	Provider          Provider
	OpenAIAPIKey      string
	OpenAIModel       string
	ClaudeAPIKey      string
	ClaudeModel       string
	GeminiProjectID   string
	GeminiLocation    string
	GeminiModel       string
	GeminiCredentials string
}

// ChatOptions は LLM 呼び出しのオプション
type ChatOptions struct {
	Temperature     *float64
	EnableWebSearch bool
}

// Client は LLM クライアントのインターフェース
type Client interface {
	Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	ChatWithOptions(ctx context.Context, systemPrompt, userPrompt string, opts ChatOptions) (string, error)
}

// NewClient は設定に応じた LLM クライアントを生成する
func NewClient(cfg ClientConfig) (Client, error) {
	switch cfg.Provider {
	case ProviderOpenAI:
		return newOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIModel), nil
	case ProviderClaude:
		return newClaudeClient(cfg.ClaudeAPIKey, cfg.ClaudeModel), nil
	case ProviderGemini:
		return newGeminiClient(cfg.GeminiProjectID, cfg.GeminiLocation, cfg.GeminiModel, cfg.GeminiCredentials)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}
