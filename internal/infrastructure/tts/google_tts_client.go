package tts

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
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
