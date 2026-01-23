package audio

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDurationMs(t *testing.T) {
	// ffprobe が利用可能かチェック
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available, skipping test")
	}

	t.Run("空のデータの場合は 0 を返す", func(t *testing.T) {
		result := GetDurationMs([]byte{})

		assert.Equal(t, 0, result)
	})

	t.Run("無効なデータの場合は 0 を返す", func(t *testing.T) {
		result := GetDurationMs([]byte("invalid audio data"))

		assert.Equal(t, 0, result)
	})

	t.Run("nil データの場合は 0 を返す", func(t *testing.T) {
		result := GetDurationMs(nil)

		assert.Equal(t, 0, result)
	})
}

func TestGetDurationMsE(t *testing.T) {
	// ffprobe が利用可能かチェック
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available, skipping test")
	}

	t.Run("空のデータの場合はエラーを返す", func(t *testing.T) {
		result, err := GetDurationMsE([]byte{})

		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("無効なデータの場合はエラーを返す", func(t *testing.T) {
		result, err := GetDurationMsE([]byte("invalid audio data"))

		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("nil データの場合はエラーを返す", func(t *testing.T) {
		result, err := GetDurationMsE(nil)

		assert.Error(t, err)
		assert.Equal(t, 0, result)
	})
}
