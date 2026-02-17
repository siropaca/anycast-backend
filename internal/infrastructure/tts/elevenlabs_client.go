package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

const (
	elevenLabsBaseURL = "https://api.elevenlabs.io"
	// Text-to-Dialogue API のデフォルト設定
	elevenLabsDialogueModelID   = "eleven_v3"
	elevenLabsDialogueLanguage  = "ja"
	elevenLabsDialogueOutputFmt = "mp3_44100_128"
	elevenLabsOutputFormat      = "mp3"
	elevenLabsTTSModelID        = "eleven_multilingual_v2"
	elevenLabsTTSOutputFmt      = "mp3_44100_128"
	elevenLabsAPITimeout        = 5 * 60 // 5 分（秒）
)

// elevenLabsTTSClient は ElevenLabs API を使った TTS クライアント
type elevenLabsTTSClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewElevenLabsTTSClient は ElevenLabs API を使った TTS クライアントを作成する
func NewElevenLabsTTSClient(apiKey string) Client {
	return &elevenLabsTTSClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: elevenLabsAPITimeout * 1e9, // ナノ秒に変換
		},
	}
}

// dialogueInput は Text-to-Dialogue API のリクエスト内の各入力要素
type dialogueInput struct {
	Text    string `json:"text"`
	VoiceID string `json:"voice_id"`
}

// dialogueRequest は Text-to-Dialogue API のリクエストボディ
type dialogueRequest struct {
	Inputs       []dialogueInput `json:"inputs"`
	ModelID      string          `json:"model_id"`
	LanguageCode string          `json:"language_code"`
}

// ttsRequest は Text-to-Speech API のリクエストボディ
type ttsRequest struct {
	Text         string `json:"text"`
	ModelID      string `json:"model_id"`
	LanguageCode string `json:"language_code"`
}

// Synthesize はテキストから音声を合成する（シングルスピーカー）
func (c *elevenLabsTTSClient) Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) (*SynthesisResult, error) {
	log := logger.FromContext(ctx)

	// emotion がある場合は [emotion] 形式でテキストの先頭に付加
	synthesisText := text
	if emotion != nil && *emotion != "" {
		synthesisText = fmt.Sprintf("[%s] %s", *emotion, text)
	}

	log.Debug("ElevenLabs TTS input", "text", synthesisText, "voiceID", voiceID)

	reqBody := ttsRequest{
		Text:         synthesisText,
		ModelID:      elevenLabsTTSModelID,
		LanguageCode: elevenLabsDialogueLanguage,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("リクエストの作成に失敗しました").WithError(err)
	}

	url := fmt.Sprintf("%s/v1/text-to-speech/%s?output_format=%s", elevenLabsBaseURL, voiceID, elevenLabsTTSOutputFmt)
	audioData, err := c.doRequest(ctx, url, body)
	if err != nil {
		log.Error("ElevenLabs TTS API error", "error", err, "voiceID", voiceID)
		return nil, err
	}

	log.Debug("ElevenLabs TTS synthesis succeeded", "audio_size", len(audioData))
	return &SynthesisResult{
		Data:   audioData,
		Format: elevenLabsOutputFormat,
	}, nil
}

// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する（マルチスピーカー）
func (c *elevenLabsTTSClient) SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig, voiceStyle *string) (*SynthesisResult, error) {
	log := logger.FromContext(ctx)

	if len(turns) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のターンが指定されていません")
	}

	if len(voiceConfigs) == 0 {
		return nil, apperror.ErrValidation.WithMessage("複数話者合成用のボイス設定が指定されていません")
	}

	// speaker alias → voice_id のマップを構築
	aliasToVoiceID := make(map[string]string, len(voiceConfigs))
	for _, vc := range voiceConfigs {
		aliasToVoiceID[vc.SpeakerAlias] = vc.VoiceID
	}

	// turns を ElevenLabs の inputs 配列に変換
	inputs := make([]dialogueInput, 0, len(turns))
	for _, turn := range turns {
		voiceID, ok := aliasToVoiceID[turn.Speaker]
		if !ok {
			return nil, apperror.ErrValidation.WithMessage(fmt.Sprintf("話者 %q の voice_id が見つかりません", turn.Speaker))
		}

		// emotion がある場合は [emotion] 形式でテキストの先頭に付加
		text := turn.Text
		if turn.Emotion != nil && *turn.Emotion != "" {
			text = fmt.Sprintf("[%s] %s", *turn.Emotion, turn.Text)
		}

		inputs = append(inputs, dialogueInput{
			Text:    text,
			VoiceID: voiceID,
		})
	}

	log.Debug("ElevenLabs Text-to-Dialogue input", "inputs_count", len(inputs))

	reqBody := dialogueRequest{
		Inputs:       inputs,
		ModelID:      elevenLabsDialogueModelID,
		LanguageCode: elevenLabsDialogueLanguage,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("リクエストの作成に失敗しました").WithError(err)
	}

	url := fmt.Sprintf("%s/v1/text-to-dialogue?output_format=%s", elevenLabsBaseURL, elevenLabsDialogueOutputFmt)
	audioData, err := c.doRequest(ctx, url, body)
	if err != nil {
		log.Error("ElevenLabs Text-to-Dialogue API error", "error", err)
		return nil, err
	}

	log.Debug("ElevenLabs Text-to-Dialogue synthesis succeeded", "audio_size", len(audioData))
	return &SynthesisResult{
		Data:   audioData,
		Format: elevenLabsOutputFormat,
	}, nil
}

// doRequest は ElevenLabs API に POST リクエストを送信し、レスポンスボディを返す
func (c *elevenLabsTTSClient) doRequest(ctx context.Context, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("HTTPリクエストの作成に失敗しました").WithError(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperror.ErrGenerationFailed.WithMessage("ElevenLabs API への接続に失敗しました").WithError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("レスポンスの読み取りに失敗しました").WithError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperror.ErrGenerationFailed.WithMessage(
			fmt.Sprintf("ElevenLabs API エラー (status=%d): %s", resp.StatusCode, string(respBody)),
		)
	}

	return respBody, nil
}
