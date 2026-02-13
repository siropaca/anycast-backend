package imagegen

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	// OpenAI 画像生成デフォルトモデル
	openAIImageGenDefaultModel = "gpt-image-1"
	// OpenAI 画像生成 API タイムアウト
	openAIImageGenTimeout = 120 * time.Second
)

// openAIImageGenClient は OpenAI Images API を使った画像生成クライアント
type openAIImageGenClient struct {
	client openai.Client
	model  openai.ImageModel
}

// NewOpenAIClient は OpenAI Images API を使った画像生成クライアントを作成する
//
// @param apiKey - OpenAI API キー
// @param model - 使用するモデル名（空文字の場合はデフォルト: gpt-image-1）
func NewOpenAIClient(apiKey, model string) Client {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithRequestTimeout(openAIImageGenTimeout),
	)

	m := openai.ImageModel(openAIImageGenDefaultModel)
	if model != "" {
		m = openai.ImageModel(model)
	}

	return &openAIImageGenClient{
		client: client,
		model:  m,
	}
}

// Generate はテキストプロンプトから画像を生成する
func (c *openAIImageGenClient) Generate(ctx context.Context, prompt string) (*GenerateResult, error) {
	log := logger.FromContext(ctx)

	log.Debug("OpenAI image generation input", "prompt", prompt, "model", c.model)

	// gpt-image-1 はデフォルトで b64_json を返すため response_format は指定しない
	resp, err := c.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Prompt: prompt,
		Model:  c.model,
		N:      openai.Int(1),
		Size:   openai.ImageGenerateParamsSize1024x1024,
	})
	if err != nil {
		log.Error("OpenAI image generation API error", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("画像生成に失敗しました").WithError(err)
	}

	if len(resp.Data) == 0 {
		log.Error("OpenAI image generation returned no data")
		return nil, apperror.ErrGenerationFailed.WithMessage("画像データが空です")
	}

	imageBytes, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		log.Error("failed to decode base64 image data", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("画像データのデコードに失敗しました").WithError(err)
	}

	log.Debug("OpenAI image generation succeeded", "image_size", len(imageBytes))
	return &GenerateResult{
		Data:     imageBytes,
		MimeType: "image/png",
	}, nil
}
