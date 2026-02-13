package imagegen

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	// Gemini 画像生成用モデル名（Vertex AI）
	geminiImageGenModelName = "gemini-2.5-flash-image"
)

// geminiImageGenClient は Gemini API を使った画像生成クライアント
type geminiImageGenClient struct {
	client *genai.Client
}

// NewGeminiClient は Gemini API を使った画像生成クライアントを作成する
// Vertex AI バックエンドを使用する
func NewGeminiClient(ctx context.Context, projectID, location, credentialsJSON string) (Client, error) {
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

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("gemini 画像生成クライアントの作成に失敗しました: %w", err)
	}

	return &geminiImageGenClient{
		client: client,
	}, nil
}

// Generate はテキストプロンプトから画像を生成する
func (c *geminiImageGenClient) Generate(ctx context.Context, prompt string) (*GenerateResult, error) {
	log := logger.FromContext(ctx)

	log.Debug("Gemini image generation input", "prompt", prompt)

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}

	resp, err := c.client.Models.GenerateContent(ctx, geminiImageGenModelName, genai.Text(prompt), config)
	if err != nil {
		log.Error("Gemini image generation API error", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("画像生成に失敗しました").WithError(err)
	}

	// 画像データを取得
	result, err := extractImageFromResponse(resp)
	if err != nil {
		log.Error("failed to extract Gemini image data", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("画像データの取得に失敗しました").WithError(err)
	}

	log.Debug("Gemini image generation succeeded", "image_size", len(result.Data), "mime_type", result.MimeType)
	return result, nil
}

// extractImageFromResponse はレスポンスから画像データを取得する
func extractImageFromResponse(resp *genai.GenerateContentResponse) (*GenerateResult, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("レスポンスが空です")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("コンテンツが空です")
	}

	for _, part := range candidate.Content.Parts {
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			return &GenerateResult{
				Data:     part.InlineData.Data,
				MimeType: part.InlineData.MIMEType,
			}, nil
		}
	}

	return nil, fmt.Errorf("画像データが見つかりません")
}
