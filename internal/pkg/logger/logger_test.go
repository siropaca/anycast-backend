package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/siropaca/anycast-backend/internal/config"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		env  config.Env
	}{
		{
			name: "本番環境で初期化",
			env:  config.EnvProduction,
		},
		{
			name: "開発環境で初期化",
			env:  config.EnvDevelopment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// リセット
			defaultLogger = nil

			Init(tt.env)

			assert.NotNil(t, defaultLogger)
			assert.Equal(t, defaultLogger, slog.Default())
		})
	}
}

func TestDefault(t *testing.T) {
	t.Run("未初期化の場合は自動で初期化される", func(t *testing.T) {
		// リセット
		defaultLogger = nil

		logger := Default()

		require.NotNil(t, logger)
		assert.Equal(t, defaultLogger, logger)
	})

	t.Run("初期化済みの場合はそのロガーを返す", func(t *testing.T) {
		Init(config.EnvDevelopment)
		expected := defaultLogger

		logger := Default()

		assert.Equal(t, expected, logger)
	})
}

func TestWithContext(t *testing.T) {
	t.Run("コンテキストにロガーを設定できる", func(t *testing.T) {
		ctx := context.Background()
		logger := slog.Default()

		newCtx := WithContext(ctx, logger)

		assert.NotEqual(t, ctx, newCtx)
		// コンテキストからロガーを取得できることを確認
		retrieved := newCtx.Value(ctxKey{})
		assert.Equal(t, logger, retrieved)
	})
}

func TestFromContext(t *testing.T) {
	t.Run("コンテキストからロガーを取得できる", func(t *testing.T) {
		logger := slog.Default()
		ctx := WithContext(context.Background(), logger)

		result := FromContext(ctx)

		assert.Equal(t, logger, result)
	})

	t.Run("コンテキストにロガーがない場合はデフォルトを返す", func(t *testing.T) {
		Init(config.EnvDevelopment)
		ctx := context.Background()

		result := FromContext(ctx)

		assert.Equal(t, defaultLogger, result)
	})
}
