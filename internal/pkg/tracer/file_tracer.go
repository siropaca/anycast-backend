package tracer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// fileTracer はファイルにトレースデータを出力するトレーサー（file モード用）
type fileTracer struct {
	dir     string
	entries map[string][]entry
}

// newFileTracer は fileTracer を生成する
func newFileTracer(episodeTitle string) *fileTracer {
	dir := filepath.Join("tmp", "traces", sanitizeTitle(episodeTitle))
	return &fileTracer{
		dir:     dir,
		entries: make(map[string][]entry),
	}
}

func (t *fileTracer) Trace(phase, section, data string) {
	t.entries[phase] = append(t.entries[phase], entry{section: section, data: data})
}

func (t *fileTracer) Flush(phase string) {
	log := logger.Default()

	entries, ok := t.entries[phase]
	if !ok || len(entries) == 0 {
		return
	}

	if err := os.MkdirAll(t.dir, 0o755); err != nil {
		log.Error("failed to create trace directory", "dir", t.dir, "error", err)
		return
	}

	content := buildMarkdown(phase, entries)
	filePath := filepath.Join(t.dir, phase+".md")

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		log.Error("failed to write trace file", "path", filePath, "error", err)
		return
	}

	log.Debug(fmt.Sprintf("trace: %s written to %s", phase, filePath))

	delete(t.entries, phase)
}

// phaseLabels は Phase 識別子に対応する表示名
var phaseLabels = map[string]string{
	"phase1": "Phase 1: ブリーフ正規化",
	"phase2": "Phase 2: 素材+アウトライン生成",
	"phase3": "Phase 3: 台本ドラフト生成",
	"phase4": "Phase 4: QA 検証+パッチ修正",
}

// buildMarkdown はトレースデータから Markdown 形式のコンテンツを構築する
func buildMarkdown(phase string, entries []entry) string {
	var sb strings.Builder

	label, ok := phaseLabels[phase]
	if !ok {
		label = phase
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", label))

	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("## %s\n\n", e.section))
		formatted := formatData(e.data)
		if isJSON(formatted) {
			sb.WriteString("```json\n")
		} else {
			sb.WriteString("```\n")
		}
		sb.WriteString(formatted)
		if !strings.HasSuffix(formatted, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	return sb.String()
}

// isJSON は文字列が JSON 形式かどうかを判定する
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// formatData はデータを見やすい形式に整形する
//
// JSON データの場合はインデント付きに変換し、それ以外はそのまま返す
func formatData(data string) string {
	if !isJSON(data) {
		return data
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return data
	}

	formatted, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return data
	}

	return string(formatted)
}

// unsafeCharsRegexp はファイル名に使用できない文字にマッチする正規表現
var unsafeCharsRegexp = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// sanitizeTitle はエピソードタイトルをディレクトリ名として安全な文字列に変換する
func sanitizeTitle(title string) string {
	s := strings.TrimSpace(title)
	s = unsafeCharsRegexp.ReplaceAllString(s, "_")

	if s == "" {
		return "untitled"
	}

	// 長すぎる場合は切り詰める
	if len(s) > 100 {
		s = s[:100]
	}

	return s
}
