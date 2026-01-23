package script

import (
	"strings"
)

// FormatLine は ScriptLine をテキスト形式に変換するための情報
type FormatLine struct {
	SpeakerName string  // 話者名
	Text        string  // セリフ
	Emotion     *string // 感情・喋り方（オプション）
}

// Format は FormatLine のスライスをテキスト形式に変換する
//
// 出力フォーマット:
//
//	話者名: [感情] セリフ
func Format(lines []FormatLine) string {
	var sb strings.Builder

	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(line.SpeakerName)
		sb.WriteString(": ")
		if line.Emotion != nil && *line.Emotion != "" {
			sb.WriteString("[")
			sb.WriteString(*line.Emotion)
			sb.WriteString("] ")
		}
		sb.WriteString(line.Text)
	}

	return sb.String()
}
