package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/siropaca/anycast-backend/internal/config"
)

var defaultLogger *slog.Logger

// 開発環境用のシンプルなハンドラー
// メッセージのみを色付きで表示する
type devHandler struct {
	level slog.Level
}

// ログレベルに応じた色を返す
func levelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "\033[31m" // 赤
	case level >= slog.LevelWarn:
		return "\033[33m" // 黄
	case level >= slog.LevelInfo:
		return "\033[32m" // 緑
	default:
		return "\033[36m" // シアン（Debug）
	}
}

func (h *devHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *devHandler) Handle(_ context.Context, r slog.Record) error {
	color := levelColor(r.Level)
	reset := "\033[0m"

	// 属性を文字列に変換
	var attrs string
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		return true
	})

	fmt.Printf("%s[%s]%s %s%s\n", color, r.Level.String(), reset, r.Message, attrs)
	return nil
}

func (h *devHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *devHandler) WithGroup(_ string) slog.Handler {
	return h
}

// ロガーを初期化する
func Init(env config.Env) {
	var handler slog.Handler

	if env == config.EnvProduction {
		opts := &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelWarn,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = &devHandler{level: slog.LevelDebug}
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// デフォルトのロガーを返す
func Default() *slog.Logger {
	if defaultLogger == nil {
		Init(config.EnvDevelopment)
	}
	return defaultLogger
}

type ctxKey struct{}

// コンテキストにロガーを設定する
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// コンテキストからロガーを取得する
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return Default()
}
