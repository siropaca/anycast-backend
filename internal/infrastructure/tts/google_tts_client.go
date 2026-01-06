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
	"github.com/siropaca/anycast-backend/internal/model"
)

const (
	// リトライ回数
	maxRetries = 3
	// デフォルト言語コード
	defaultLanguageCode = "ja-JP"
)

// TTS クライアントのインターフェース
type Client interface {
	Synthesize(ctx context.Context, text, voiceID string, gender model.Gender) ([]byte, error)
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
func (c *googleTTSClient) Synthesize(ctx context.Context, text, voiceID string, gender model.Gender) ([]byte, error) {
	log := logger.FromContext(ctx)

	// 自然な発声を促すためのSSMLラップ
	// <speak>タグで囲み、わずかな「間」や「ピッチ」の調整を加える土台を作ります
	ssml := fmt.Sprintf("<speak>%s</speak>", text)

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Ssml{
				Ssml: ssml,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: defaultLanguageCode,
			Name:         voiceID,
			SsmlGender:   toSsmlVoiceGender(gender),
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:    texttospeechpb.AudioEncoding_MP3,
			SampleRateHertz:  24000, // MP3 の最大サポート値（ポッドキャスト品質として十分）
			Pitch:            0.0,
			SpeakingRate:     1.05,                               // 日本語の場合、やや速めにすると自然に聞こえる
			EffectsProfileId: []string{"headphone-class-device"}, // ヘッドホン/イヤホン向けに最適化
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
