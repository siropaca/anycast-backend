package script

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		text            string
		allowedSpeakers []string
		allowedSfx      []string
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
			name:            "正常系: SILENCE を含む",
			text:            "太郎: こんにちは\n__SILENCE__: 800\n花子: やあ",
			allowedSpeakers: []string{"太郎", "花子"},
			wantLines:       3,
			wantErrors:      0,
		},
		{
			name:            "正常系: SFX を含む（チェックなし）",
			text:            "太郎: こんにちは\n__SFX__: chime\n花子: やあ",
			allowedSpeakers: []string{"太郎", "花子"},
			allowedSfx:      nil,
			wantLines:       3,
			wantErrors:      0,
		},
		{
			name:            "正常系: SFX を含む（許可リストあり）",
			text:            "__SFX__: chime",
			allowedSpeakers: []string{},
			allowedSfx:      []string{"chime", "bell"},
			wantLines:       1,
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
			name:            "エラー: SILENCE の値が不正",
			text:            "__SILENCE__: abc",
			allowedSpeakers: []string{},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "エラー: SILENCE の値が負数",
			text:            "__SILENCE__: -100",
			allowedSpeakers: []string{},
			wantLines:       0,
			wantErrors:      1,
		},
		{
			name:            "エラー: SFX が許可リストにない",
			text:            "__SFX__: unknown",
			allowedSpeakers: []string{},
			allowedSfx:      []string{"chime", "bell"},
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
			result := Parse(tt.text, tt.allowedSpeakers, tt.allowedSfx)

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

	result := Parse(text, allowedSpeakers, nil)

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

func TestParse_LineTypes(t *testing.T) {
	text := `太郎: こんにちは
__SILENCE__: 500
__SFX__: chime
花子: やあ`
	allowedSpeakers := []string{"太郎", "花子"}

	result := Parse(text, allowedSpeakers, nil)

	if len(result.Lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(result.Lines))
	}

	// 1行目: speech
	if result.Lines[0].LineType != LineTypeSpeech {
		t.Errorf("expected LineTypeSpeech, got %s", result.Lines[0].LineType)
	}
	if result.Lines[0].SpeakerName != "太郎" {
		t.Errorf("expected speaker '太郎', got '%s'", result.Lines[0].SpeakerName)
	}

	// 2行目: silence
	if result.Lines[1].LineType != LineTypeSilence {
		t.Errorf("expected LineTypeSilence, got %s", result.Lines[1].LineType)
	}
	if result.Lines[1].DurationMs != 500 {
		t.Errorf("expected durationMs 500, got %d", result.Lines[1].DurationMs)
	}

	// 3行目: sfx
	if result.Lines[2].LineType != LineTypeSfx {
		t.Errorf("expected LineTypeSfx, got %s", result.Lines[2].LineType)
	}
	if result.Lines[2].SfxName != "chime" {
		t.Errorf("expected sfxName 'chime', got '%s'", result.Lines[2].SfxName)
	}

	// 4行目: speech
	if result.Lines[3].LineType != LineTypeSpeech {
		t.Errorf("expected LineTypeSpeech, got %s", result.Lines[3].LineType)
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
