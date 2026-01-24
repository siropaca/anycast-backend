package tts

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/genai"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	// Gemini TTS 用モデル名（Vertex AI）
	geminiAPITTSModelName = "gemini-2.5-pro-tts"
	// デフォルト言語コード
	geminiDefaultLanguageCode = "ja-JP"
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
	// Synthesize はテキストから音声を合成する（シングルスピーカー）
	Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error)
	// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する
	SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) ([]byte, error)
	// SupportsLongForm は長い台本を分割せずに処理できるかどうかを返す
	SupportsLongForm() bool
	// OutputFormat は音声出力フォーマットを返す（"mp3" または "pcm"）
	OutputFormat() string
}

// geminiTTSClient は Gemini API を使った TTS クライアント
// 32k token のコンテキストウィンドウをサポートし、長い台本でも一貫した音声を生成できる
type geminiTTSClient struct {
	client *genai.Client
}

// NewGeminiTTSClient は Gemini API を使った TTS クライアントを作成する
// Vertex AI バックエンドを使用する
func NewGeminiTTSClient(ctx context.Context, projectID, location, credentialsJSON string) (Client, error) {
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
		return nil, fmt.Errorf("gemini TTS クライアントの作成に失敗しました: %w", err)
	}

	return &geminiTTSClient{
		client: client,
	}, nil
}

// Synthesize はテキストから音声を合成する（シングルスピーカー）
func (c *geminiTTSClient) Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) ([]byte, error) {
	log := logger.FromContext(ctx)

	// emotion がある場合は [emotion] 形式でテキストの先頭に付加
	synthesisText := text
	if emotion != nil && *emotion != "" {
		synthesisText = fmt.Sprintf("[%s] %s", *emotion, text)
	}

	log.Debug("Gemini TTS 入力", "text", synthesisText, "voiceID", voiceID)

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			LanguageCode: geminiDefaultLanguageCode,
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: voiceID,
				},
			},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, geminiAPITTSModelName, genai.Text(synthesisText), config)
	if err != nil {
		log.Error("Gemini TTS API エラー", "error", err, "voiceID", voiceID)
		return nil, apperror.ErrGenerationFailed.WithMessage("音声合成に失敗しました").WithError(err)
	}

	// 音声データを取得
	audioData, err := extractAudioFromResponse(resp)
	if err != nil {
		log.Error("Gemini TTS 音声データの取得に失敗しました", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("音声データの取得に失敗しました").WithError(err)
	}

	log.Debug("Gemini TTS 音声合成に成功しました", "audio_size", len(audioData))
	return audioData, nil
}

// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する
func (c *geminiTTSClient) SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) ([]byte, error) {
	log := logger.FromContext(ctx)

	if len(turns) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のターンが指定されていません")
	}

	if len(voiceConfigs) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のボイス設定が指定されていません")
	}

	// プロンプトを構築（マルチスピーカー形式）
	var promptBuilder strings.Builder

	// voiceStyle が指定されている場合は先頭に追加
	if voiceStyle != nil && *voiceStyle != "" {
		promptBuilder.WriteString(*voiceStyle)
		promptBuilder.WriteString("\n\n")
	}

	// 台本を構築
	for _, turn := range turns {
		text := turn.Text
		if turn.Emotion != nil && *turn.Emotion != "" {
			text = fmt.Sprintf("[%s] %s", *turn.Emotion, turn.Text)
		}
		// 行間にポーズを追加
		text += " [medium pause]"

		promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", turn.Speaker, text))
	}

	prompt := promptBuilder.String()
	log.Debug("Gemini TTS マルチスピーカープロンプト", "prompt_length", len(prompt), "turns_count", len(turns))

	// SpeakerVoiceConfigs を構築
	speakerVoiceConfigs := make([]*genai.SpeakerVoiceConfig, len(voiceConfigs))
	for i, vc := range voiceConfigs {
		speakerVoiceConfigs[i] = &genai.SpeakerVoiceConfig{
			Speaker: vc.SpeakerAlias,
			VoiceConfig: &genai.VoiceConfig{
				PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
					VoiceName: vc.VoiceID,
				},
			},
		}
	}

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"AUDIO"},
		SpeechConfig: &genai.SpeechConfig{
			LanguageCode: geminiDefaultLanguageCode,
			MultiSpeakerVoiceConfig: &genai.MultiSpeakerVoiceConfig{
				SpeakerVoiceConfigs: speakerVoiceConfigs,
			},
		},
	}

	resp, err := c.client.Models.GenerateContent(ctx, geminiAPITTSModelName, genai.Text(prompt), config)
	if err != nil {
		log.Error("Gemini TTS マルチスピーカー API エラー", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声合成に失敗しました").WithError(err)
	}

	// 音声データを取得
	audioData, err := extractAudioFromResponse(resp)
	if err != nil {
		log.Error("Gemini TTS マルチスピーカー音声データの取得に失敗しました", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声データの取得に失敗しました").WithError(err)
	}

	log.Debug("Gemini TTS マルチスピーカー音声合成に成功しました", "audio_size", len(audioData))
	return audioData, nil
}

// extractAudioFromResponse はレスポンスから音声データを取得する
func extractAudioFromResponse(resp *genai.GenerateContentResponse) ([]byte, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("レスポンスが空です")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("コンテンツが空です")
	}

	for _, part := range candidate.Content.Parts {
		if part.InlineData != nil && len(part.InlineData.Data) > 0 {
			return part.InlineData.Data, nil
		}
	}

	return nil, fmt.Errorf("音声データが見つかりません")
}

// Close はクライアントを閉じる
func (c *geminiTTSClient) Close() error {
	// genai.Client には Close メソッドがない場合がある
	return nil
}

// SupportsLongForm は長い台本を分割せずに処理できるかどうかを返す
// Gemini TTS は 32k token をサポートするため true
func (c *geminiTTSClient) SupportsLongForm() bool {
	return true
}

// OutputFormat は音声出力フォーマットを返す
// Gemini TTS は PCM 16bit 24kHz を返す
func (c *geminiTTSClient) OutputFormat() string {
	return "pcm"
}
