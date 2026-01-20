package script

import (
	"testing"
)

func TestFormat(t *testing.T) {
	emotion := "嬉しそうに"

	tests := []struct {
		name  string
		lines []FormatLine
		want  string
	}{
		{
			name: "正常系: 基本的な会話",
			lines: []FormatLine{
				{SpeakerName: "太郎", Text: "こんにちは"},
				{SpeakerName: "花子", Text: "やあ"},
			},
			want: "太郎: こんにちは\n花子: やあ",
		},
		{
			name: "正常系: 感情付き",
			lines: []FormatLine{
				{SpeakerName: "太郎", Text: "こんにちは", Emotion: &emotion},
			},
			want: "太郎: [嬉しそうに] こんにちは",
		},
		{
			name:  "正常系: 空のスライス",
			lines: []FormatLine{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.lines)
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormat_EdgeCases(t *testing.T) {
	emptyEmotion := ""

	tests := []struct {
		name  string
		lines []FormatLine
		want  string
	}{
		{
			name: "話者名が空",
			lines: []FormatLine{
				{SpeakerName: "", Text: "こんにちは"},
			},
			want: ": こんにちは",
		},
		{
			name: "セリフが空",
			lines: []FormatLine{
				{SpeakerName: "太郎", Text: ""},
			},
			want: "太郎: ",
		},
		{
			name: "感情が空文字列",
			lines: []FormatLine{
				{SpeakerName: "太郎", Text: "こんにちは", Emotion: &emptyEmotion},
			},
			want: "太郎: こんにちは",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.lines)
			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormat_RoundTrip(t *testing.T) {
	// フォーマット → パース → フォーマットで同じ結果になることを確認
	emotion := "嬉しそうに"
	original := []FormatLine{
		{SpeakerName: "太郎", Text: "こんにちは", Emotion: &emotion},
		{SpeakerName: "花子", Text: "やあ"},
	}

	// フォーマット
	text := Format(original)

	// パース
	parseResult := Parse(text, []string{"太郎", "花子"})

	if parseResult.HasErrors() {
		t.Fatalf("unexpected parse errors: %v", parseResult.Errors)
	}

	// 再度フォーマット用に変換
	formatted := make([]FormatLine, len(parseResult.Lines))
	for i, line := range parseResult.Lines {
		formatted[i] = FormatLine(line)
	}

	// 再フォーマット
	text2 := Format(formatted)

	if text != text2 {
		t.Errorf("round trip failed:\noriginal: %q\nafter: %q", text, text2)
	}
}
