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

// SynthesisResult は TTS 合成結果を表す
type SynthesisResult struct {
	Data       []byte // 音声バイナリデータ
	Format     string // "pcm" or "mp3"
	SampleRate int    // PCM の場合のサンプルレート（MP3 の場合は 0）
}

// Client は TTS クライアントのインターフェース
type Client interface {
	// Synthesize はテキストから音声を合成する（シングルスピーカー）
	Synthesize(ctx context.Context, text string, emotion *string, voiceID string, gender model.Gender) (*SynthesisResult, error)
	// SynthesizeMultiSpeaker は複数話者のテキストから音声を合成する（マルチスピーカー）
	SynthesizeMultiSpeaker(ctx context.Context, turns []SpeakerTurn, voiceConfigs []SpeakerVoiceConfig) (*SynthesisResult, error)
}
