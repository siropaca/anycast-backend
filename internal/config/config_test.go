package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/logger"
)

func TestLoad(t *testing.T) {
	t.Run("環境変数が未設定の場合はデフォルト値を使用する", func(t *testing.T) {
		t.Setenv("PORT", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("APP_ENV", "")

		cfg := Load()

		assert.Equal(t, "8081", cfg.Port)
		assert.Equal(t, "", cfg.DatabaseURL)
		assert.Equal(t, logger.Env("development"), cfg.AppEnv)
	})

	t.Run("環境変数が設定されている場合はその値を使用する", func(t *testing.T) {
		t.Setenv("PORT", "9000")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
		t.Setenv("APP_ENV", "production")

		cfg := Load()

		assert.Equal(t, "9000", cfg.Port)
		assert.Equal(t, "postgres://localhost:5432/test", cfg.DatabaseURL)
		assert.Equal(t, logger.Env("production"), cfg.AppEnv)
	})

	t.Run("PORT のみ設定した場合は他はデフォルト値を使用する", func(t *testing.T) {
		t.Setenv("PORT", "3000")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("APP_ENV", "")

		cfg := Load()

		assert.Equal(t, "3000", cfg.Port)
		assert.Equal(t, "", cfg.DatabaseURL)
		assert.Equal(t, logger.Env("development"), cfg.AppEnv)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("環境変数が設定されている場合はその値を返す", func(t *testing.T) {
		t.Setenv("TEST_VAR", "test_value")

		result := getEnv("TEST_VAR", "default")

		assert.Equal(t, "test_value", result)
	})

	t.Run("環境変数が未設定の場合はデフォルト値を返す", func(t *testing.T) {
		t.Setenv("TEST_VAR", "")

		result := getEnv("TEST_VAR", "default")

		assert.Equal(t, "default", result)
	})

	t.Run("存在しない環境変数の場合はデフォルト値を返す", func(t *testing.T) {
		result := getEnv("NON_EXISTENT_VAR_12345", "default")

		assert.Equal(t, "default", result)
	})
}
