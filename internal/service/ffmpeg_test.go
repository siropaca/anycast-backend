package service

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "正の整数",
			input:    15.0,
			expected: "15.000",
		},
		{
			name:     "負の数",
			input:    -15.0,
			expected: "-15.000",
		},
		{
			name:     "小数点以下3桁",
			input:    1.234,
			expected: "1.234",
		},
		{
			name:     "小数点以下3桁以上は切り捨て",
			input:    1.23456789,
			expected: "1.235",
		},
		{
			name:     "ゼロ",
			input:    0,
			expected: "0.000",
		},
		{
			name:     "小さい数",
			input:    0.001,
			expected: "0.001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFloat(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewFFmpegService(t *testing.T) {
	t.Run("FFmpegService を作成できる", func(t *testing.T) {
		service := NewFFmpegService()

		assert.NotNil(t, service)
	})
}

func TestFFmpegService_MixAudioWithBGM(t *testing.T) {
	// ffmpeg が利用可能かチェック
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping test")
	}

	service := NewFFmpegService()
	ctx := context.Background()

	t.Run("空の音声データの場合はエラーを返す", func(t *testing.T) {
		params := MixParams{
			VoiceData:       []byte{},
			BGMData:         []byte{},
			VoiceDurationMs: 1000,
			BGMVolumeDB:     -20.0,
			FadeOutMs:       3000,
			PaddingStartMs:  1000,
			PaddingEndMs:    1000,
		}

		_, err := service.MixAudioWithBGM(ctx, params)

		assert.Error(t, err)
	})

	t.Run("無効な音声データの場合はエラーを返す", func(t *testing.T) {
		params := MixParams{
			VoiceData:       []byte("invalid audio data"),
			BGMData:         []byte("invalid bgm data"),
			VoiceDurationMs: 1000,
			BGMVolumeDB:     -20.0,
			FadeOutMs:       3000,
			PaddingStartMs:  1000,
			PaddingEndMs:    1000,
		}

		_, err := service.MixAudioWithBGM(ctx, params)

		assert.Error(t, err)
	})
}

func TestMixParams(t *testing.T) {
	t.Run("MixParams にデフォルト値を設定できる", func(t *testing.T) {
		params := MixParams{
			VoiceData:       []byte("voice"),
			BGMData:         []byte("bgm"),
			VoiceDurationMs: 5000,
			BGMVolumeDB:     -20.0,
			FadeOutMs:       3000,
			PaddingStartMs:  1000,
			PaddingEndMs:    1000,
		}

		assert.Equal(t, 5000, params.VoiceDurationMs)
		assert.Equal(t, -20.0, params.BGMVolumeDB)
		assert.Equal(t, 3000, params.FadeOutMs)
		assert.Equal(t, 1000, params.PaddingStartMs)
		assert.Equal(t, 1000, params.PaddingEndMs)
	})
}
