package script

import (
	"fmt"
	"strings"
)

// ScriptLine をテキスト形式に変換するための情報
type FormatLine struct {
	LineType    LineType // 行の種別
	SpeakerName string   // 話者名（speech 時のみ）
	Text        string   // セリフ（speech 時のみ）
	Emotion     *string  // 感情・喋り方（speech 時のみ、オプション）
	DurationMs  int      // 無音の長さ（silence 時のみ）
	SfxName     string   // 効果音名（sfx 時のみ）
}

// FormatLine のスライスをテキスト形式に変換する
//
// 出力フォーマット:
//
//	話者名: [感情] セリフ
//	__SILENCE__: ミリ秒
//	__SFX__: 効果音名
func Format(lines []FormatLine) string {
	var sb strings.Builder

	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}

		switch line.LineType {
		case LineTypeSpeech:
			sb.WriteString(line.SpeakerName)
			sb.WriteString(": ")
			if line.Emotion != nil && *line.Emotion != "" {
				sb.WriteString("[")
				sb.WriteString(*line.Emotion)
				sb.WriteString("] ")
			}
			sb.WriteString(line.Text)

		case LineTypeSilence:
			sb.WriteString(SilenceKeyword)
			sb.WriteString(": ")
			sb.WriteString(fmt.Sprintf("%d", line.DurationMs))

		case LineTypeSfx:
			sb.WriteString(SfxKeyword)
			sb.WriteString(": ")
			sb.WriteString(line.SfxName)
		}
	}

	return sb.String()
}
