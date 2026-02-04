package llm

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"

	"github.com/siropaca/anycast-backend/internal/pkg/prompt"
)

const (
	// Gemini デフォルトモデル
	geminiDefaultModel = "gemini-2.5-flash"
)

type geminiClient struct {
	client *genai.Client
	model  string
}

// newGeminiClient は Gemini（Vertex AI）クライアントを生成する
func newGeminiClient(projectID, location, model, credentialsJSON string) (Client, error) {
	config := &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  projectID,
		Location: location,
	}

	// credentialsJSON が指定されている場合は使用
	if credentialsJSON != "" {
		creds, err := credentials.DetectDefault(&credentials.DetectOptions{
			Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
			CredentialsJSON: []byte(credentialsJSON),
		})
		if err != nil {
			return nil, fmt.Errorf("認証情報の読み込みに失敗しました: %w", err)
		}
		config.Credentials = creds
	}

	client, err := genai.NewClient(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("Gemini LLM クライアントの作成に失敗しました: %w", err)
	}

	m := geminiDefaultModel
	if model != "" {
		m = model
	}

	return &geminiClient{
		client: client,
		model:  m,
	}, nil
}

// Chat はシステムプロンプトとユーザープロンプトを使って LLM と対話する
func (c *geminiClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return retryWithBackoff(ctx, "Gemini", func() (string, error) {
		resp, err := c.client.Models.GenerateContent(ctx,
			c.model,
			genai.Text(prompt.Compress(userPrompt)),
			&genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{{Text: prompt.Compress(systemPrompt)}},
				},
				Temperature: genai.Ptr(float32(defaultTemperature)),
			},
		)
		if err != nil {
			return "", err
		}

		return resp.Text(), nil
	})
}
