package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	t.Run("トークンを生成できる", func(t *testing.T) {
		token, err := Generate()
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		// 32 バイトを base64url エンコードすると 43 文字になる
		assert.Len(t, token, 43)
	})

	t.Run("生成されるトークンはユニークである", func(t *testing.T) {
		tokens := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := Generate()
			require.NoError(t, err)
			assert.False(t, tokens[token], "重複するトークンが生成されました: %s", token)
			tokens[token] = true
		}
	})
}
