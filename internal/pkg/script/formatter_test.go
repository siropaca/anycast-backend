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
			name: "正常系: speech のみ",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは"},
				{LineType: LineTypeSpeech, SpeakerName: "花子", Text: "やあ"},
			},
			want: "太郎: こんにちは\n花子: やあ",
		},
		{
			name: "正常系: 感情付き",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは", Emotion: &emotion},
			},
			want: "太郎: [嬉しそうに] こんにちは",
		},
		{
			name: "正常系: silence",
			lines: []FormatLine{
				{LineType: LineTypeSilence, DurationMs: 800},
			},
			want: "__SILENCE__: 800",
		},
		{
			name: "正常系: sfx",
			lines: []FormatLine{
				{LineType: LineTypeSfx, SfxName: "chime"},
			},
			want: "__SFX__: chime",
		},
		{
			name: "正常系: 複合",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは"},
				{LineType: LineTypeSilence, DurationMs: 500},
				{LineType: LineTypeSfx, SfxName: "chime"},
				{LineType: LineTypeSpeech, SpeakerName: "花子", Text: "やあ"},
			},
			want: "太郎: こんにちは\n__SILENCE__: 500\n__SFX__: chime\n花子: やあ",
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
			name: "不明な LineType は出力されない",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは"},
				{LineType: LineType("unknown"), SpeakerName: "謎", Text: "テスト"},
				{LineType: LineTypeSpeech, SpeakerName: "花子", Text: "やあ"},
			},
			want: "太郎: こんにちは\n\n花子: やあ",
		},
		{
			name: "speech で話者名が空",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "", Text: "こんにちは"},
			},
			want: ": こんにちは",
		},
		{
			name: "speech でセリフが空",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: ""},
			},
			want: "太郎: ",
		},
		{
			name: "speech で感情が空文字列",
			lines: []FormatLine{
				{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは", Emotion: &emptyEmotion},
			},
			want: "太郎: こんにちは",
		},
		{
			name: "silence で DurationMs が 0",
			lines: []FormatLine{
				{LineType: LineTypeSilence, DurationMs: 0},
			},
			want: "__SILENCE__: 0",
		},
		{
			name: "sfx で SfxName が空",
			lines: []FormatLine{
				{LineType: LineTypeSfx, SfxName: ""},
			},
			want: "__SFX__: ",
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
		{LineType: LineTypeSpeech, SpeakerName: "太郎", Text: "こんにちは", Emotion: &emotion},
		{LineType: LineTypeSilence, DurationMs: 800},
		{LineType: LineTypeSfx, SfxName: "chime"},
		{LineType: LineTypeSpeech, SpeakerName: "花子", Text: "やあ"},
	}

	// フォーマット
	text := Format(original)

	// パース
	parseResult := Parse(text, []string{"太郎", "花子"}, nil)

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
