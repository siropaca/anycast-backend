package tts

import (
	"context"
	"fmt"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

const (
	// Gemini TTS モデル名
	geminiTTSModelName = "gemini-2.5-pro-tts"
	// デフォルト言語コード
	defaultLanguageCode = "ja-JP"
	// リトライ回数
	maxRetries = 3
)

// TTS クライアントのインターフェース
type Client interface {
	Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error)
}

type googleTTSClient struct {
	client *texttospeech.Client
}

// Google TTS クライアントを作成する
func NewGoogleTTSClient(ctx context.Context, credentialsJSON string) (Client, error) {
	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	}

	client, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client: %w", err)
	}

	return &googleTTSClient{
		client: client,
	}, nil
}

// テキストから音声を合成する
// Gemini-TTS を使用し、emotion を Prompt フィールドでスタイル指示として渡す
func (c *googleTTSClient) Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error) {
	log := logger.FromContext(ctx)

	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{
			Text: text,
		},
	}

	// emotion がある場合は Prompt フィールドに設定
	if emotion != nil && *emotion != "" {
		input.Prompt = emotion
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: input,
		Voice: &texttospeechpb.VoiceSelectionParams{
			ModelName:    geminiTTSModelName,
			LanguageCode: defaultLanguageCode,
			Name:         voiceID,
			SsmlGender:   toSsmlVoiceGender(gender),
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("synthesizing speech", "attempt", attempt, "text_length", len(text))

		resp, err := c.client.SynthesizeSpeech(ctx, req)
		if err != nil {
			lastErr = err
			log.Warn(fmt.Sprintf("tts api error: attempt=%d, voiceID=%s, error=%v", attempt, voiceID, err))

			// 最後のリトライでなければ待機して再試行
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("tts api failed after retries", "error", err, "voiceID", voiceID)
			return nil, apperror.ErrGenerationFailed.WithMessage("Failed to synthesize speech").WithError(err)
		}

		if len(resp.AudioContent) == 0 {
			lastErr = fmt.Errorf("empty audio content in response")
			log.Warn("empty audio content in tts response", "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("tts returned empty audio after retries")
			return nil, apperror.ErrGenerationFailed.WithMessage("Failed to synthesize speech: empty audio")
		}

		log.Debug("speech synthesized successfully", "audio_size", len(resp.AudioContent))
		return resp.AudioContent, nil
	}

	return nil, apperror.ErrGenerationFailed.WithMessage("Failed to synthesize speech").WithError(lastErr)
}

// Gender を SsmlVoiceGender に変換する
func toSsmlVoiceGender(gender model.Gender) texttospeechpb.SsmlVoiceGender {
	switch gender {
	case model.GenderMale:
		return texttospeechpb.SsmlVoiceGender_MALE
	case model.GenderFemale:
		return texttospeechpb.SsmlVoiceGender_FEMALE
	case model.GenderNeutral:
		return texttospeechpb.SsmlVoiceGender_NEUTRAL
	default:
		return texttospeechpb.SsmlVoiceGender_SSML_VOICE_GENDER_UNSPECIFIED
	}
}

// クライアントを閉じる
func (c *googleTTSClient) Close() error {
	return c.client.Close()
}
