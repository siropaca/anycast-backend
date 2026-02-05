package tracer

// Mode はトレースモードを表す型
type Mode string

const (
	ModeNone Mode = "none"
	ModeLog  Mode = "log"
	ModeFile Mode = "file"
)

// Tracer は台本生成の各 Phase のデータをトレースするインターフェース
type Tracer interface {
	// Trace は phase 名とセクション名をキーにデータを蓄積する
	Trace(phase string, section string, data string)

	// Flush は指定 phase の蓄積データをまとめて出力する
	Flush(phase string)
}

// entry はトレースデータの1エントリを表す
type entry struct {
	section string
	data    string
}

// New は指定されたモードに応じた Tracer を生成する
func New(mode Mode, episodeTitle string) Tracer {
	switch mode {
	case ModeLog:
		return newLogTracer()
	case ModeFile:
		return newFileTracer(episodeTitle)
	default:
		return &noopTracer{}
	}
}
