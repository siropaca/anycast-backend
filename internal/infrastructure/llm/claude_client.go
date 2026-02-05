package llm

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/siropaca/anycast-backend/internal/pkg/prompt"
)

const (
	// Claude デフォルトモデル
	claudeDefaultModel = anthropic.ModelClaudeSonnet4_20250514
	// Claude の最大出力トークン数
	claudeMaxTokens = 8192
)

type claudeClient struct {
	client anthropic.Client
	model  anthropic.Model
}

// newClaudeClient は Claude（Anthropic）クライアントを生成する
func newClaudeClient(apiKey, model string) Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	m := claudeDefaultModel
	if model != "" {
		m = anthropic.Model(model)
	}

	return &claudeClient{
		client: client,
		model:  m,
	}
}

// Chat はシステムプロンプトとユーザープロンプトを使って LLM と対話する
func (c *claudeClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.ChatWithOptions(ctx, systemPrompt, userPrompt, ChatOptions{})
}

// ChatWithOptions はオプション付きで LLM と対話する
func (c *claudeClient) ChatWithOptions(ctx context.Context, systemPrompt, userPrompt string, opts ChatOptions) (string, error) {
	temp := defaultTemperature
	if opts.Temperature != nil {
		temp = *opts.Temperature
	}

	params := anthropic.MessageNewParams{
		MaxTokens: claudeMaxTokens,
		Model:     c.model,
		System: []anthropic.TextBlockParam{
			{Text: prompt.Compress(systemPrompt)},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewTextBlock(prompt.Compress(userPrompt)),
			),
		},
		Temperature: anthropic.Float(temp),
	}

	if opts.EnableWebSearch {
		params.Tools = []anthropic.ToolUnionParam{
			{
				OfWebSearchTool20250305: &anthropic.WebSearchTool20250305Param{},
			},
		}
	}

	retryName := "Claude"
	if opts.EnableWebSearch {
		retryName = "Claude(WebSearch)"
	}

	return retryWithBackoff(ctx, retryName, func() (string, error) {
		message, err := c.client.Messages.New(ctx, params)
		if err != nil {
			return "", err
		}

		// レスポンスからテキストを抽出
		var result string
		for _, block := range message.Content {
			text := block.AsText()
			if text.Text != "" {
				result += text.Text
			}
		}

		return result, nil
	})
}
