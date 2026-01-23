package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/prompt"
)

const (
	// 使用するモデル
	defaultModel = openai.ChatModelGPT5_2
	// 生成時の Temperature（0.0〜2.0）
	defaultTemperature = 0.7
	// API タイムアウト
	defaultTimeout = 120 * time.Second
	// リトライ回数
	maxRetries = 3
)

// Client は LLM クライアントのインターフェース
type Client interface {
	GenerateScript(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type openAIClient struct {
	client openai.Client
}

// NewOpenAIClient は OpenAI クライアントを作成する
func NewOpenAIClient(apiKey string) Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(defaultTimeout),
	)

	return &openAIClient{
		client: client,
	}
}

// GenerateScript は台本を生成する
func (c *openAIClient) GenerateScript(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	log := logger.FromContext(ctx)

	log.Debug("GenerateScript 呼び出し", "systemPrompt", systemPrompt)
	log.Debug("GenerateScript 呼び出し", "userPrompt", userPrompt)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("台本を生成中", "attempt", attempt)

		resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: defaultModel,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt.Compress(systemPrompt)),
				openai.UserMessage(prompt.Compress(userPrompt)),
			},
			Temperature: openai.Float(defaultTemperature),
		})

		if err != nil {
			lastErr = err
			log.Warn("OpenAI API エラー", "attempt", attempt, "error", err)

			// 最後のリトライでなければ待機して再試行
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("OpenAI API がリトライ後も失敗しました", "error", err)
			return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました").WithError(err)
		}

		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices in response")
			log.Warn("OpenAI レスポンスに選択肢がありません", "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("OpenAI がリトライ後も選択肢を返しませんでした")
			return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました: レスポンスがありません")
		}

		content := resp.Choices[0].Message.Content
		log.Debug("台本生成に成功しました", "content_length", len(content))

		return content, nil
	}

	return "", apperror.ErrGenerationFailed.WithMessage("台本の生成に失敗しました").WithError(lastErr)
}
