package audio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSilencePCM(t *testing.T) {
	t.Run("200ms の無音を生成する (24kHz, mono, s16le)", func(t *testing.T) {
		result := GenerateSilencePCM(200, 24000, 1, 2)

		// 24000 * 200/1000 * 1 * 2 = 9600 bytes
		assert.Equal(t, 9600, len(result))
		// すべてゼロであることを確認
		for i, b := range result {
			if b != 0 {
				t.Fatalf("expected zero at index %d, got %d", i, b)
			}
		}
	})

	t.Run("0ms の場合は空のスライスを返す", func(t *testing.T) {
		result := GenerateSilencePCM(0, 24000, 1, 2)

		assert.Empty(t, result)
	})

	t.Run("ステレオ (2ch) の場合はバイト数が2倍になる", func(t *testing.T) {
		mono := GenerateSilencePCM(100, 24000, 1, 2)
		stereo := GenerateSilencePCM(100, 24000, 2, 2)

		assert.Equal(t, len(mono)*2, len(stereo))
	})
}

func TestConcatPCM(t *testing.T) {
	t.Run("複数のセグメントを結合する", func(t *testing.T) {
		seg1 := []byte{1, 2, 3, 4}
		seg2 := []byte{5, 6, 7, 8}
		seg3 := []byte{9, 10}

		result := ConcatPCM([][]byte{seg1, seg2, seg3})

		assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, result)
	})

	t.Run("単一セグメントの場合はそのまま返す", func(t *testing.T) {
		seg := []byte{1, 2, 3}

		result := ConcatPCM([][]byte{seg})

		assert.Equal(t, []byte{1, 2, 3}, result)
	})

	t.Run("空のスライスの場合は空を返す", func(t *testing.T) {
		result := ConcatPCM([][]byte{})

		assert.Empty(t, result)
	})

	t.Run("nil の場合は空を返す", func(t *testing.T) {
		result := ConcatPCM(nil)

		assert.Empty(t, result)
	})

	t.Run("空のセグメントを含む場合も正しく結合する", func(t *testing.T) {
		seg1 := []byte{1, 2}
		seg2 := []byte{}
		seg3 := []byte{3, 4}

		result := ConcatPCM([][]byte{seg1, seg2, seg3})

		assert.Equal(t, []byte{1, 2, 3, 4}, result)
	})
}
