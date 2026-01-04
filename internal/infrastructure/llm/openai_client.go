package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/prompt"
)

const (
	// API タイムアウト
	defaultTimeout = 120 * time.Second
	// リトライ回数
	maxRetries = 3
	// 生成時の Temperature（0.0〜2.0）
	defaultTemperature = 0.7
)

// LLM クライアントのインターフェース
type Client interface {
	GenerateScript(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type openAIClient struct {
	client openai.Client
}

// OpenAI クライアントを作成する
func NewOpenAIClient(apiKey string) Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(defaultTimeout),
	)

	return &openAIClient{
		client: client,
	}
}

// 台本を生成する
func (c *openAIClient) GenerateScript(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	log := logger.FromContext(ctx)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("generating script", "attempt", attempt)

		resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT4o,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt.Compress(systemPrompt)),
				openai.UserMessage(prompt.Compress(userPrompt)),
			},
			Temperature: openai.Float(defaultTemperature),
		})

		if err != nil {
			lastErr = err
			log.Warn("openai api error", "attempt", attempt, "error", err)

			// 最後のリトライでなければ待機して再試行
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("openai api failed after retries", "error", err)
			return "", apperror.ErrGenerationFailed.WithMessage("Failed to generate script").WithError(err)
		}

		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			log.Warn("no choices in openai response", "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("openai returned no choices after retries")
			return "", apperror.ErrGenerationFailed.WithMessage("Failed to generate script: no response")
		}

		content := resp.Choices[0].Message.Content
		log.Debug("script generated successfully", "content_length", len(content))

		return content, nil
	}

	return "", apperror.ErrGenerationFailed.WithMessage("Failed to generate script").WithError(lastErr)
}
