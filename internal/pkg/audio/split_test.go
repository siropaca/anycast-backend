package audio

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSilenceDetectOutput(t *testing.T) {
	t.Run("標準的な silencedetect 出力をパースする", func(t *testing.T) {
		stderr := `[silencedetect @ 0x1234] silence_start: 1.5
[silencedetect @ 0x1234] silence_end: 2.0 | silence_duration: 0.5
[silencedetect @ 0x1234] silence_start: 4.2
[silencedetect @ 0x1234] silence_end: 4.8 | silence_duration: 0.6`

		intervals := parseSilenceDetectOutput(stderr)

		require.Len(t, intervals, 2)
		assert.InDelta(t, 1.5, intervals[0].StartSec, 0.001)
		assert.InDelta(t, 2.0, intervals[0].EndSec, 0.001)
		assert.InDelta(t, 4.2, intervals[1].StartSec, 0.001)
		assert.InDelta(t, 4.8, intervals[1].EndSec, 0.001)
	})

	t.Run("無音区間がない場合は空スライスを返す", func(t *testing.T) {
		stderr := `size=N/A time=00:00:05.00 bitrate=N/A speed=100x`

		intervals := parseSilenceDetectOutput(stderr)

		assert.Empty(t, intervals)
	})

	t.Run("start のみで end がない場合は無視する", func(t *testing.T) {
		stderr := `[silencedetect @ 0x1234] silence_start: 1.5
[silencedetect @ 0x1234] silence_start: 3.0`

		intervals := parseSilenceDetectOutput(stderr)

		assert.Empty(t, intervals)
	})

	t.Run("小数点以下の精度が異なる値を正しくパースする", func(t *testing.T) {
		stderr := `[silencedetect @ 0x1234] silence_start: 0.123456
[silencedetect @ 0x1234] silence_end: 0.654321 | silence_duration: 0.530865`

		intervals := parseSilenceDetectOutput(stderr)

		require.Len(t, intervals, 1)
		assert.InDelta(t, 0.123456, intervals[0].StartSec, 0.000001)
		assert.InDelta(t, 0.654321, intervals[0].EndSec, 0.000001)
	})

	t.Run("空文字列の場合は空スライスを返す", func(t *testing.T) {
		intervals := parseSilenceDetectOutput("")

		assert.Empty(t, intervals)
	})
}

func TestSelectTopSilenceIntervals(t *testing.T) {
	t.Run("無音区間を長さ順で上位 n 個を時系列順に返す", func(t *testing.T) {
		intervals := []SilenceInterval{
			{StartSec: 1.0, EndSec: 1.3}, // 0.3s
			{StartSec: 3.0, EndSec: 4.0}, // 1.0s (longest)
			{StartSec: 5.0, EndSec: 5.5}, // 0.5s
			{StartSec: 7.0, EndSec: 7.8}, // 0.8s
		}

		// 上位 2 個を選択 → 3.0-4.0 (1.0s) と 7.0-7.8 (0.8s) が時系列順で返る
		result := selectTopSilenceIntervals(intervals, 2)

		require.Len(t, result, 2)
		assert.InDelta(t, 3.0, result[0].StartSec, 0.001)
		assert.InDelta(t, 4.0, result[0].EndSec, 0.001)
		assert.InDelta(t, 7.0, result[1].StartSec, 0.001)
		assert.InDelta(t, 7.8, result[1].EndSec, 0.001)
	})

	t.Run("n が区間数以上の場合はそのまま返す", func(t *testing.T) {
		intervals := []SilenceInterval{
			{StartSec: 1.0, EndSec: 1.5},
			{StartSec: 3.0, EndSec: 3.5},
		}

		result := selectTopSilenceIntervals(intervals, 5)

		assert.Len(t, result, 2)
	})

	t.Run("n=1 の場合は最長の区間だけ返す", func(t *testing.T) {
		intervals := []SilenceInterval{
			{StartSec: 1.0, EndSec: 1.2}, // 0.2s
			{StartSec: 3.0, EndSec: 4.5}, // 1.5s
			{StartSec: 6.0, EndSec: 6.3}, // 0.3s
		}

		result := selectTopSilenceIntervals(intervals, 1)

		require.Len(t, result, 1)
		assert.InDelta(t, 3.0, result[0].StartSec, 0.001)
	})
}

func TestDetectSilenceIntervals(t *testing.T) {
	// ffmpeg が利用可能かチェック
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping test")
	}

	config := PCMSplitConfig{
		SampleRate:     24000,
		Channels:       1,
		BytesPerSample: 2,
		NoiseDB:        -30.0,
		MinSilenceSec:  0.3,
	}

	t.Run("空データの場合はエラーを返す", func(t *testing.T) {
		_, err := DetectSilenceIntervals([]byte{}, config)

		assert.Error(t, err)
	})

	t.Run("無音を含む PCM から無音区間を検出する", func(t *testing.T) {
		bytesPerSec := 24000 * 1 * 2

		// 500ms のノイズ + 500ms の無音 + 500ms のノイズ
		noise1 := make([]byte, bytesPerSec/2)
		silence := make([]byte, bytesPerSec/2)
		noise2 := make([]byte, bytesPerSec/2)

		// s16le で十分大きな振幅のノイズを生成（値 0x2000 = 8192）
		for i := 0; i+1 < len(noise1); i += 2 {
			noise1[i] = 0x00
			noise1[i+1] = 0x20
		}
		for i := 0; i+1 < len(noise2); i += 2 {
			noise2[i] = 0x00
			noise2[i+1] = 0x20
		}

		pcmData := ConcatPCM([][]byte{noise1, silence, noise2})

		intervals, err := DetectSilenceIntervals(pcmData, config)

		require.NoError(t, err)
		require.NotEmpty(t, intervals)
		// 無音区間は 0.5s 付近から始まるはず
		assert.InDelta(t, 0.5, intervals[0].StartSec, 0.15)
	})

	t.Run("無音がない PCM では空スライスを返す", func(t *testing.T) {
		bytesPerSec := 24000 * 1 * 2
		noise := make([]byte, bytesPerSec) // 1秒のノイズ

		// s16le で十分大きな振幅のノイズを生成
		for i := 0; i+1 < len(noise); i += 2 {
			noise[i] = 0x00
			noise[i+1] = 0x20
		}

		intervals, err := DetectSilenceIntervals(noise, config)

		require.NoError(t, err)
		assert.Empty(t, intervals)
	})
}

func TestSplitPCMBySilence(t *testing.T) {
	// ffmpeg が利用可能かチェック
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping test")
	}

	config := PCMSplitConfig{
		SampleRate:     24000,
		Channels:       1,
		BytesPerSample: 2,
		NoiseDB:        -30.0,
		MinSilenceSec:  0.3,
	}

	t.Run("空データの場合はエラーを返す", func(t *testing.T) {
		_, err := SplitPCMBySilence([]byte{}, config)

		assert.Error(t, err)
	})

	t.Run("無音のみの PCM を分割する", func(t *testing.T) {
		// 2秒の無音データ（24kHz, mono, s16le）
		silenceData := GenerateSilencePCM(2000, 24000, 1, 2)

		segments, err := SplitPCMBySilence(silenceData, config)

		// 無音のみなので1セグメント（silencedetect の挙動次第でバリエーションあり）
		require.NoError(t, err)
		assert.NotEmpty(t, segments)

		// 全セグメントを合計したバイト数が元データと一致する
		totalLen := 0
		for _, seg := range segments {
			totalLen += len(seg)
		}
		assert.Equal(t, len(silenceData), totalLen)
	})

	t.Run("音声+無音+音声の PCM を2セグメントに分割する", func(t *testing.T) {
		bytesPerSec := 24000 * 1 * 2 // 48000 bytes/sec

		// 500ms のノイズ + 500ms の無音 + 500ms のノイズ
		noise1 := make([]byte, bytesPerSec/2)  // 500ms
		silence := make([]byte, bytesPerSec/2) // 500ms
		noise2 := make([]byte, bytesPerSec/2)  // 500ms

		// ノイズデータを生成（交互パターンで非ゼロ）
		for i := range noise1 {
			if i%2 == 0 {
				noise1[i] = 0x40 // ある程度大きな値
			}
		}
		for i := range noise2 {
			if i%2 == 0 {
				noise2[i] = 0x40
			}
		}

		pcmData := ConcatPCM([][]byte{noise1, silence, noise2})

		segments, err := SplitPCMBySilence(pcmData, config)

		require.NoError(t, err)
		assert.Equal(t, 2, len(segments))

		// 全セグメントを合計したバイト数が元データと一致する
		totalLen := 0
		for _, seg := range segments {
			totalLen += len(seg)
		}
		assert.Equal(t, len(pcmData), totalLen)
	})
}
