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

	// デフォルトの音声スタイルプロンプト
	// ポッドキャストとして聞き取りやすいベーススタイルを定義する
	defaultVoiceStyle = "ポッドキャスト番組の収録です。落ち着いたテンポでゆっくり話しつつも、自然な抑揚と感情を込めて、友達と雑談するように楽しく語ってください。"

	// Gemini TTS の出力フォーマット
	geminiOutputFormat     = "pcm"
	geminiOutputSampleRate = 24000
)

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
func (c *geminiTTSClient) Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) (*SynthesisResult, error) {
	log := logger.FromContext(ctx)

	// emotion がある場合は [emotion] 形式でテキストの先頭に付加
	synthesisText := text
	if emotion != nil && *emotion != "" {
		synthesisText = fmt.Sprintf("[%s] %s", *emotion, text)
	}

	log.Debug("Gemini TTS input", "text", synthesisText, "voiceID", voiceID)

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
		log.Error("Gemini TTS API error", "error", err, "voiceID", voiceID)
		return nil, apperror.ErrGenerationFailed.WithMessage("音声合成に失敗しました").WithError(err)
	}

	// 音声データを取得
	audioData, err := extractAudioFromResponse(resp)
	if err != nil {
		log.Error("failed to extract Gemini TTS audio data", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("音声データの取得に失敗しました").WithError(err)
	}

	log.Debug("Gemini TTS synthesis succeeded", "audio_size", len(audioData))
	return &SynthesisResult{
		Data:       audioData,
		Format:     geminiOutputFormat,
		SampleRate: geminiOutputSampleRate,
	}, nil
}

// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する（マルチスピーカー）
func (c *geminiTTSClient) SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) (*SynthesisResult, error) {
	log := logger.FromContext(ctx)

	if len(turns) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のターンが指定されていません")
	}

	if len(voiceConfigs) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のボイス設定が指定されていません")
	}

	// 台本を構築（マルチスピーカー形式）
	var promptBuilder strings.Builder

	// デフォルトの音声スタイルを先頭に追加
	promptBuilder.WriteString(defaultVoiceStyle)

	// ユーザー指定の voiceStyle がある場合は追記
	if voiceStyle != nil && *voiceStyle != "" {
		promptBuilder.WriteString("\n")
		promptBuilder.WriteString(*voiceStyle)
	}

	promptBuilder.WriteString("\n\n")

	for _, turn := range turns {
		text := turn.Text
		if turn.Emotion != nil && *turn.Emotion != "" {
			text = fmt.Sprintf("[%s] %s", *turn.Emotion, turn.Text)
		}

		promptBuilder.WriteString(fmt.Sprintf("%s: %s\n", turn.Speaker, text))
	}

	if voiceStyle != nil && *voiceStyle != "" {
		log.Debug("Gemini TTS voice style", "voice_style", *voiceStyle)
	}

	prompt := promptBuilder.String()
	log.Debug("Gemini TTS script", "prompt", prompt)
	log.Debug("starting Gemini TTS multi-speaker processing", "prompt_length", len(prompt), "turns_count", len(turns))

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
		log.Error("Gemini TTS multi-speaker API error", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声合成に失敗しました").WithError(err)
	}

	// 音声データを取得
	audioData, err := extractAudioFromResponse(resp)
	if err != nil {
		log.Error("failed to extract Gemini TTS multi-speaker audio data", "error", err)
		return nil, apperror.ErrGenerationFailed.WithMessage("複数話者の音声データの取得に失敗しました").WithError(err)
	}

	log.Debug("Gemini TTS multi-speaker synthesis succeeded", "audio_size", len(audioData))
	return &SynthesisResult{
		Data:       audioData,
		Format:     geminiOutputFormat,
		SampleRate: geminiOutputSampleRate,
	}, nil
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
