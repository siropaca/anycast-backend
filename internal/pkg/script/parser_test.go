package script

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		allowedSpeakers []string
		wantLines       int
		wantErrors      int
	}{
		{
			name: "正常系: 基本的な会話",
			text: `太郎: こんにちは！
花子: やあ、元気？
太郎: 元気だよ！`,
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       3,
			wantErrors:      0,
		},
		{
			name: "正常系: 感情付きの会話",
			text: `太郎: [嬉しそうに] こんにちは！
花子: [驚いて] あ、太郎くん！`,
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       2,
			wantErrors:      0,
		},
		{
			name:            "正常系: 空行を含む",
			text:            "太郎: こんにちは\n\n花子: やあ",
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       2,
			wantErrors:      0,
		},
		{
			name:            "エラー: 不明な話者",
			text:            "三郎: こんにちは",
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "エラー: コロンなし",
			text:            "これはセリフじゃない",
			allowedSpeakers: []string{"太郎"},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "エラー: セリフが空",
			text:            "太郎: ",
			allowedSpeakers: []string{"太郎"},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "エラー: 感情のみでセリフが空",
			text:            "太郎: [嬉しそうに]",
			allowedSpeakers: []string{"太郎"},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "複合: 正常行とエラー行が混在",
			text:            "太郎: こんにちは\n三郎: やあ\n花子: 元気？",
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       2,
			wantErrors:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.text, tt.allowedSpeakers)

			if len(result.Lines) != tt.wantLines {
				t.Errorf("Parse() lines = %d, want %d", len(result.Lines), tt.wantLines)
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Parse() errors = %d, want %d", len(result.Errors), tt.wantErrors)
			}
		})
	}
}

func TestParse_EmotionExtraction(t *testing.T) {
	text := `太郎: [嬉しそうに] こんにちは！
花子: やあ、元気？`
	allowedSpeakers := []string{"太郎", "花子"}

	result := Parse(text, allowedSpeakers)

	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result.Lines))
	}

	// 1行目: 感情あり
	if result.Lines[0].Emotion == nil {
		t.Error("expected emotion for first line")
	} else if *result.Lines[0].Emotion != "嬉しそうに" {
		t.Errorf("expected emotion '嬉しそうに', got '%s'", *result.Lines[0].Emotion)
	}
	if result.Lines[0].Text != "こんにちは！" {
		t.Errorf("expected text 'こんにちは！', got '%s'", result.Lines[0].Text)
	}

	// 2行目: 感情なし
	if result.Lines[1].Emotion != nil {
		t.Error("expected no emotion for second line")
	}
	if result.Lines[1].Text != "やあ、元気？" {
		t.Errorf("expected text 'やあ、元気？', got '%s'", result.Lines[1].Text)
	}
}

func TestParse_SpeakerExtraction(t *testing.T) {
	text := `太郎: こんにちは
花子: やあ`
	allowedSpeakers := []string{"太郎", "花子"}

	result := Parse(text, allowedSpeakers)

	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(result.Lines))
	}

	if result.Lines[0].SpeakerName != "太郎" {
		t.Errorf("expected speaker '太郎', got '%s'", result.Lines[0].SpeakerName)
	}

	if result.Lines[1].SpeakerName != "花子" {
		t.Errorf("expected speaker '花子', got '%s'", result.Lines[1].SpeakerName)
	}
}

func TestCountNonEmptyLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "通常の行",
			text: "太郎: こんにちは\n花子: やあ",
			want: 2,
		},
		{
			name: "空行を含む",
			text: "太郎: こんにちは\n\n\n花子: やあ",
			want: 2,
		},
		{
			name: "空文字列",
			text: "",
			want: 0,
		},
		{
			name: "空白のみの行を含む",
			text: "太郎: こんにちは\n   \n花子: やあ",
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountNonEmptyLines(tt.text); got != tt.want {
				t.Errorf("CountNonEmptyLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseResult_HasErrors(t *testing.T) {
	tests := []struct {
		name       string
		result     ParseResult
		wantErrors bool
	}{
		{
			name:       "エラーなし",
			result:     ParseResult{Lines: []ParsedLine{{SpeakerName: "太郎", Text: "こんにちは"}}, Errors: []ParseError{}},
			wantErrors: false,
		},
		{
			name:       "エラーあり",
			result:     ParseResult{Lines: []ParsedLine{}, Errors: []ParseError{{Line: 1, Reason: "test"}}},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasErrors(); got != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", got, tt.wantErrors)
			}
		})
	}
}

func TestStripEmotionTags(t *testing.T) {
	text := "ケンタ: [嬉しそうに] こんにちは\nミホ: [笑いながら] よろしく\nケンタ: 普通のセリフ"
	result := StripEmotionTags(text)
	expected := "ケンタ: こんにちは\nミホ: よろしく\nケンタ: 普通のセリフ"
	if result != expected {
		t.Errorf("StripEmotionTags:\ngot:  %q\nwant: %q", result, expected)
	}
}

func TestCapEmotionTags(t *testing.T) {
	text := "ケンタ: [嬉しそうに] こんにちは\nミホ: [笑いながら] よろしく\nケンタ: [興奮して] テスト\nミホ: [呆れて] もう一つ\nケンタ: 普通のセリフ"
	result := CapEmotionTags(text, 2)
	// Count emotion tags in result
	lines := splitLines(result)
	count := 0
	for _, l := range lines {
		if emotionRegex.MatchString(afterColon(l)) {
			count++
		}
	}
	if count != 2 {
		t.Errorf("CapEmotionTags(maxTags=2): got %d emotion tags, want 2\nresult: %s", count, result)
	}
}

func splitLines(s string) []string {
	result := []string{}
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			result = append(result, l)
		}
	}
	return result
}

func afterColon(s string) string {
	idx := strings.Index(s, ":")
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(s[idx+1:])
}
