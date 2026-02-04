package tracer

import (
	"fmt"
	"strings"

	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// logTracer は slog の Debug レベルでトレース出力するトレーサー（log モード用）
type logTracer struct {
	entries map[string][]entry
}

// newLogTracer は logTracer を生成する
func newLogTracer() *logTracer {
	return &logTracer{
		entries: make(map[string][]entry),
	}
}

func (t *logTracer) Trace(phase, section, data string) {
	t.entries[phase] = append(t.entries[phase], entry{section: section, data: data})
}

func (t *logTracer) Flush(phase string) {
	log := logger.Default()

	entries, ok := t.entries[phase]
	if !ok || len(entries) == 0 {
		return
	}

	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("[%s] %s", e.section, e.data))
		sb.WriteString("\n---\n")
	}

	log.Debug(fmt.Sprintf("trace: %s", phase), "data", sb.String())

	delete(t.entries, phase)
}
