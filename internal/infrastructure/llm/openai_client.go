package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"github.com/siropaca/anycast-backend/internal/pkg/prompt"
)

const (
	// OpenAI デフォルトモデル
	openAIDefaultModel = openai.ChatModelGPT5_2
	// 生成時の Temperature（0.0〜2.0）
	defaultTemperature = 0.7
	// API タイムアウト
	defaultTimeout = 120 * time.Second
)

type openAIClient struct {
	client openai.Client
	model  openai.ChatModel
}

// newOpenAIClient は OpenAI クライアントを生成する
func newOpenAIClient(apiKey, model string) Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(defaultTimeout),
	)

	m := openAIDefaultModel
	if model != "" {
		m = openai.ChatModel(model)
	}

	return &openAIClient{
		client: client,
		model:  m,
	}
}

// Chat はシステムプロンプトとユーザープロンプトを使って LLM と対話する
func (c *openAIClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.ChatWithOptions(ctx, systemPrompt, userPrompt, ChatOptions{})
}

// ChatWithOptions はオプション付きで LLM と対話する
func (c *openAIClient) ChatWithOptions(ctx context.Context, systemPrompt, userPrompt string, opts ChatOptions) (string, error) {
	if opts.EnableWebSearch {
		return c.chatWithResponsesAPI(ctx, systemPrompt, userPrompt, opts)
	}

	temp := defaultTemperature
	if opts.Temperature != nil {
		temp = *opts.Temperature
	}

	return retryWithBackoff(ctx, "OpenAI", func() (string, error) {
		resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: c.model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(prompt.Compress(systemPrompt)),
				openai.UserMessage(prompt.Compress(userPrompt)),
			},
			Temperature: openai.Float(temp),
		})
		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no choices in response")
		}

		return resp.Choices[0].Message.Content, nil
	})
}

// chatWithResponsesAPI は Responses API を使って web_search 付きで LLM と対話する
func (c *openAIClient) chatWithResponsesAPI(ctx context.Context, systemPrompt, userPrompt string, opts ChatOptions) (string, error) {
	temp := defaultTemperature
	if opts.Temperature != nil {
		temp = *opts.Temperature
	}

	return retryWithBackoff(ctx, "OpenAI(Responses)", func() (string, error) {
		resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
			Model:        string(c.model),
			Instructions: openai.String(prompt.Compress(systemPrompt)),
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(prompt.Compress(userPrompt)),
			},
			Temperature: openai.Float(temp),
			Tools: []responses.ToolUnionParam{
				{
					OfWebSearch: &responses.WebSearchToolParam{
						Type:              responses.WebSearchToolTypeWebSearch,
						SearchContextSize: responses.WebSearchToolSearchContextSizeHigh,
					},
				},
			},
		})
		if err != nil {
			return "", err
		}

		return resp.OutputText(), nil
	})
}
