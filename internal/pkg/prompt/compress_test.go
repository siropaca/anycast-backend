package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompress(t *testing.T) {
	t.Run("連続する改行を1つにまとめる", func(t *testing.T) {
		input := "line1\n\n\nline2\n\nline3"
		expected := "line1\nline2\nline3"
		assert.Equal(t, expected, Compress(input))
	})

	t.Run("行頭のスペースを除去する", func(t *testing.T) {
		input := "  line1\n    line2\n\tline3"
		expected := "line1\nline2\nline3"
		assert.Equal(t, expected, Compress(input))
	})

	t.Run("先頭と末尾の空白を除去する", func(t *testing.T) {
		input := "\n\n  hello world  \n\n"
		expected := "hello world"
		assert.Equal(t, expected, Compress(input))
	})

	t.Run("Markdownリスト形式は維持される", func(t *testing.T) {
		input := "## ルール\n- item1\n- item2"
		expected := "## ルール\n- item1\n- item2"
		assert.Equal(t, expected, Compress(input))
	})

	t.Run("行中のスペースは維持される", func(t *testing.T) {
		input := "hello   world"
		expected := "hello   world"
		assert.Equal(t, expected, Compress(input))
	})

	t.Run("複合的なケース", func(t *testing.T) {
		input := `
  ## タイトル

  - アイテム1
  - アイテム2


  説明文
`
		expected := "## タイトル\n- アイテム1\n- アイテム2\n説明文"
		assert.Equal(t, expected, Compress(input))
	})
}
