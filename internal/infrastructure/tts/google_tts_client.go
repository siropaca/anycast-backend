package tts

import (
	"bytes"
	"context"
	"fmt"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
)

const (
	// リトライ回数
	maxRetries = 3
	// デフォルト言語コード
	defaultLanguageCode = "ja-JP"
)

// TTS クライアントのインターフェース
type Client interface {
	Synthesize(ctx context.Context, text, voiceID string) ([]byte, error)
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
func (c *googleTTSClient) Synthesize(ctx context.Context, text, voiceID string) ([]byte, error) {
	log := logger.FromContext(ctx)

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: defaultLanguageCode,
			Name:         voiceID,
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

// クライアントを閉じる
func (c *googleTTSClient) Close() error {
	return c.client.Close()
}

// MP3 データから再生時間（ミリ秒）を取得する
func GetMP3DurationMs(data []byte) (int, error) {
	reader := bytes.NewReader(data)

	// MP3 ファイルのヘッダーを解析して duration を計算
	// シンプルなビットレートベースの推定を使用
	// MP3 は通常 128kbps なので、サイズから推定
	// duration (秒) = ファイルサイズ (バイト) * 8 / ビットレート (bps)
	const defaultBitrate = 128000 // 128 kbps

	fileSize := reader.Size()
	durationSeconds := float64(fileSize*8) / float64(defaultBitrate)
	durationMs := int(durationSeconds * 1000)

	return durationMs, nil
}
