package tracer

// noopTracer は何も出力しないトレーサー（none モード用）
type noopTracer struct{}

func (t *noopTracer) Trace(phase, section, data string) {}

func (t *noopTracer) Flush(phase string) {}
