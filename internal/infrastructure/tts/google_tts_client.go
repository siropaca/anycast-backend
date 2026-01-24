package tts

import (
	"context"
	"fmt"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	// Gemini TTS モデル名
	geminiTTSModelName = "gemini-2.5-pro-tts"
	// デフォルト言語コード
	defaultLanguageCode = "ja-JP"
	// リトライ回数
	maxRetries = 3
)

// SpeakerTurn は multi-speaker 合成用の話者とテキストのペア
type SpeakerTurn struct {
	Speaker string  // 話者名（キャラクター名）
	Text    string  // セリフ
	Emotion *string // 感情（オプション）
}

// SpeakerVoiceConfig は話者名と Voice ID のマッピング
type SpeakerVoiceConfig struct {
	SpeakerAlias string // 話者名（スクリプト内での名前）
	VoiceID      string // Voice の ProviderVoiceID
}

// Client は TTS クライアントのインターフェース
type Client interface {
	Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error)
	SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) ([]byte, error)
}

type googleTTSClient struct {
	client *texttospeech.Client
}

// NewGoogleTTSClient は Google TTS クライアントを作成する
func NewGoogleTTSClient(ctx context.Context, credentialsJSON string) (Client, error) {
	var opts []option.ClientOption
	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON))) //nolint:staticcheck // TODO: migrate to newer auth method
	}

	client, err := texttospeech.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("TTS クライアントの作成に失敗しました: %w", err)
	}

	return &googleTTSClient{
		client: client,
	}, nil
}

// Synthesize はテキストから音声を合成する
// Gemini-TTS を使用し、emotion がある場合は [emotion] をテキストの先頭に付加する
func (c *googleTTSClient) Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error) {
	log := logger.FromContext(ctx)

	// emotion がある場合は [emotion] 形式でテキストの先頭に付加
	synthesisText := text
	if emotion != nil && *emotion != "" {
		synthesisText = fmt.Sprintf("[%s] %s", *emotion, text)
	}

	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{
			Text: synthesisText,
		},
	}

	log.Debug("TTS 入力", "text", synthesisText, "voiceID", voiceID)

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
		log.Debug("音声を合成中", "attempt", attempt, "text_length", len(text))

		resp, err := c.client.SynthesizeSpeech(ctx, req)
		if err != nil {
			lastErr = err
			log.Warn(fmt.Sprintf("TTS API エラー: 試行=%d, voiceID=%s, error=%v", attempt, voiceID, err))

			// 最後のリトライでなければ待機して再試行
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("TTS API がリトライ後も失敗しました", "error", err, "voiceID", voiceID)
			return nil, apperror.ErrGenerationFailed.WithMessage("音声合成に失敗しました").WithError(err)
		}

		if len(resp.AudioContent) == 0 {
			lastErr = fmt.Errorf("レスポンスの音声コンテンツが空です")
			log.Warn("TTS レスポンスの音声コンテンツが空です", "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("TTS がリトライ後も空の音声を返しました")
			return nil, apperror.ErrGenerationFailed.WithMessage("音声合成に失敗しました: 音声データが空です")
		}

		log.Debug("音声合成に成功しました", "audio_size", len(resp.AudioContent))
		return resp.AudioContent, nil
	}

	return nil, apperror.ErrGenerationFailed.WithMessage("音声合成に失敗しました").WithError(lastErr)
}

// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する
// Gemini-TTS の multi-speaker 機能を使用
func (c *googleTTSClient) SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) ([]byte, error) {
	log := logger.FromContext(ctx)

	if len(turns) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のターンが指定されていません")
	}

	if len(voiceConfigs) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のボイス設定が指定されていません")
	}

	// MultiSpeakerMarkup を構築
	markupTurns := make([]*texttospeechpb.MultiSpeakerMarkup_Turn, len(turns))
	for i, turn := range turns {
		// emotion がある場合は [emotion] 形式でテキストの先頭に付加
		text := turn.Text
		if turn.Emotion != nil && *turn.Emotion != "" {
			text = fmt.Sprintf("[%s] %s", *turn.Emotion, turn.Text)
		}

		// 行の最後にポーズを追加して行間に余白を入れる
		text = text + " [medium pause]"

		markupTurns[i] = &texttospeechpb.MultiSpeakerMarkup_Turn{
			Speaker: turn.Speaker,
			Text:    text,
		}

		log.Debug("multi-speaker turn", "speaker", turn.Speaker, "text", text)
	}

	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_MultiSpeakerMarkup{
			MultiSpeakerMarkup: &texttospeechpb.MultiSpeakerMarkup{
				Turns: markupTurns,
			},
		},
	}

	// voiceStyle が指定されている場合は Prompt に設定
	// AI Studio では Style Instructions と呼ばれる
	if voiceStyle != nil && *voiceStyle != "" {
		input.Prompt = voiceStyle
		log.Debug("ボイススタイルを使用中", "voiceStyle", *voiceStyle)
	}

	// SpeakerVoiceConfigs を構築
	speakerVoiceConfigs := make([]*texttospeechpb.MultispeakerPrebuiltVoice, len(voiceConfigs))
	for i, vc := range voiceConfigs {
		speakerVoiceConfigs[i] = &texttospeechpb.MultispeakerPrebuiltVoice{
			SpeakerAlias: vc.SpeakerAlias,
			SpeakerId:    vc.VoiceID,
		}
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: input,
		Voice: &texttospeechpb.VoiceSelectionParams{
			ModelName:    geminiTTSModelName,
			LanguageCode: defaultLanguageCode,
			MultiSpeakerVoiceConfig: &texttospeechpb.MultiSpeakerVoiceConfig{
				SpeakerVoiceConfigs: speakerVoiceConfigs,
			},
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Debug("マルチスピーカー音声を合成中", "attempt", attempt, "turns_count", len(turns), "speakers_count", len(voiceConfigs))

		resp, err := c.client.SynthesizeSpeech(ctx, req)
		if err != nil {
			lastErr = err
			log.Warn(fmt.Sprintf("マルチスピーカーTTS API エラー: 試行=%d, error=%v", attempt, err))

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("マルチスピーカーTTS API がリトライ後も失敗しました", "error", err)
			return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声合成に失敗しました").WithError(err)
		}

		if len(resp.AudioContent) == 0 {
			lastErr = fmt.Errorf("レスポンスの音声コンテンツが空です")
			log.Warn("マルチスピーカーTTS レスポンスの音声コンテンツが空です", "attempt", attempt)

			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}

			log.Error("マルチスピーカーTTS がリトライ後も空の音声を返しました")
			return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声合成に失敗しました: 音声データが空です")
		}

		log.Debug("マルチスピーカー音声合成に成功しました", "audio_size", len(resp.AudioContent))
		return resp.AudioContent, nil
	}

	return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声合成に失敗しました").WithError(lastErr)
}

// toSsmlVoiceGender は Gender を SsmlVoiceGender に変換する
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

// Close はクライアントを閉じる
func (c *googleTTSClient) Close() error {
	return c.client.Close()
}
