package logger

import (
	"context"
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger

// Init はロガーを初期化する
func Init(env string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	if env == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Default はデフォルトのロガーを返す
func Default() *slog.Logger {
	if defaultLogger == nil {
		Init("development")
	}
	return defaultLogger
}

type ctxKey struct{}

// WithContext はコンテキストにロガーを設定する
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext はコンテキストからロガーを取得する
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return Default()
}
