package tracer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("none モードは noopTracer を返す", func(t *testing.T) {
		tr := New(ModeNone, "test")
		_, ok := tr.(*noopTracer)
		assert.True(t, ok)
	})

	t.Run("空文字列は noopTracer を返す", func(t *testing.T) {
		tr := New("", "test")
		_, ok := tr.(*noopTracer)
		assert.True(t, ok)
	})

	t.Run("log モードは logTracer を返す", func(t *testing.T) {
		tr := New(ModeLog, "test")
		_, ok := tr.(*logTracer)
		assert.True(t, ok)
	})

	t.Run("file モードは fileTracer を返す", func(t *testing.T) {
		tr := New(ModeFile, "test")
		_, ok := tr.(*fileTracer)
		assert.True(t, ok)
	})
}

func TestNoopTracer(t *testing.T) {
	t.Run("Trace と Flush がパニックしない", func(t *testing.T) {
		tr := &noopTracer{}
		assert.NotPanics(t, func() {
			tr.Trace("phase1", "system_prompt", "test data")
			tr.Flush("phase1")
		})
	})
}

func TestLogTracer(t *testing.T) {
	t.Run("Trace でデータを蓄積し Flush で削除される", func(t *testing.T) {
		tr := newLogTracer()

		tr.Trace("phase1", "system_prompt", "prompt data")
		tr.Trace("phase1", "response", "response data")

		assert.Len(t, tr.entries["phase1"], 2)

		tr.Flush("phase1")

		assert.Empty(t, tr.entries["phase1"])
	})

	t.Run("存在しない phase の Flush はパニックしない", func(t *testing.T) {
		tr := newLogTracer()
		assert.NotPanics(t, func() {
			tr.Flush("nonexistent")
		})
	})
}

func TestFileTracer(t *testing.T) {
	t.Run("Flush でファイルが作成される", func(t *testing.T) {
		tmpDir := t.TempDir()
		tr := &fileTracer{
			dir:     filepath.Join(tmpDir, "test-episode"),
			entries: make(map[string][]entry),
		}

		tr.Trace("phase2", "system_prompt", "system prompt content")
		tr.Trace("phase2", "user_prompt", `{"episode":{"title":"test"}}`)
		tr.Trace("phase2", "response", `{"grounding":{}}`)

		tr.Flush("phase2")

		filePath := filepath.Join(tmpDir, "test-episode", "phase2.md")
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "# Phase 2: 素材+アウトライン生成")
		assert.Contains(t, contentStr, "## system_prompt")
		assert.Contains(t, contentStr, "system prompt content")
		assert.Contains(t, contentStr, "## user_prompt")
		assert.Contains(t, contentStr, "## response")
	})

	t.Run("Flush 後にエントリが削除される", func(t *testing.T) {
		tmpDir := t.TempDir()
		tr := &fileTracer{
			dir:     filepath.Join(tmpDir, "test-episode"),
			entries: make(map[string][]entry),
		}

		tr.Trace("phase1", "brief", "brief data")
		tr.Flush("phase1")

		assert.Empty(t, tr.entries["phase1"])
	})

	t.Run("存在しない phase の Flush はパニックしない", func(t *testing.T) {
		tr := &fileTracer{
			dir:     t.TempDir(),
			entries: make(map[string][]entry),
		}
		assert.NotPanics(t, func() {
			tr.Flush("nonexistent")
		})
	})
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"JSON オブジェクト", `{"key":"value"}`, true},
		{"JSON 配列", `[1,2,3]`, true},
		{"前後に空白がある JSON", `  {"key":"value"}  `, true},
		{"プレーンテキスト", "hello world", false},
		{"空文字列", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isJSON(tt.input))
		})
	}
}

func TestFormatData(t *testing.T) {
	t.Run("JSON を整形する", func(t *testing.T) {
		input := `{"episode":{"title":"test"}}`
		result := formatData(input)
		assert.Contains(t, result, "  \"episode\"")
		assert.Contains(t, result, "    \"title\": \"test\"")
	})

	t.Run("非 JSON はそのまま返す", func(t *testing.T) {
		input := "plain text"
		assert.Equal(t, input, formatData(input))
	})

	t.Run("不正な JSON はそのまま返す", func(t *testing.T) {
		input := `{invalid json}`
		assert.Equal(t, input, formatData(input))
	})
}

func TestSanitizeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "通常のタイトル",
			input:    "AIの未来について",
			expected: "AIの未来について",
		},
		{
			name:     "不正文字を含むタイトル",
			input:    `AI<>の"未来"/test`,
			expected: "AI__の_未来__test",
		},
		{
			name:     "空文字列",
			input:    "",
			expected: "untitled",
		},
		{
			name:     "スペースのみ",
			input:    "   ",
			expected: "untitled",
		},
		{
			name:     "前後のスペースをトリム",
			input:    "  タイトル  ",
			expected: "タイトル",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildMarkdown(t *testing.T) {
	t.Run("既知の phase ラベルが使用される", func(t *testing.T) {
		entries := []entry{
			{section: "system_prompt", data: "prompt"},
		}
		result := buildMarkdown("phase2", entries)
		assert.Contains(t, result, "# Phase 2: 素材+アウトライン生成")
	})

	t.Run("未知の phase はそのまま表示される", func(t *testing.T) {
		entries := []entry{
			{section: "data", data: "test"},
		}
		result := buildMarkdown("unknown_phase", entries)
		assert.Contains(t, result, "# unknown_phase")
	})

	t.Run("JSON データが整形されて json コードブロックで囲まれる", func(t *testing.T) {
		entries := []entry{
			{section: "response", data: `{"key":"value"}`},
		}
		result := buildMarkdown("phase1", entries)
		assert.Contains(t, result, "```json\n")
		assert.Contains(t, result, "\"key\": \"value\"")
	})

	t.Run("非 JSON データはそのままコードブロックで囲まれる", func(t *testing.T) {
		entries := []entry{
			{section: "system_prompt", data: "plain text prompt"},
		}
		result := buildMarkdown("phase1", entries)
		assert.Contains(t, result, "```\nplain text prompt\n```")
	})
}
