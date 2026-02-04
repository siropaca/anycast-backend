package script

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractJSON(t *testing.T) {
	t.Run("json コードブロックから抽出", func(t *testing.T) {
		text := "以下が結果です：\n```json\n{\"key\": \"value\"}\n```\n以上です。"
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Equal(t, `{"key": "value"}`, result)
	})

	t.Run("コードブロックから抽出", func(t *testing.T) {
		text := "結果：\n```\n{\"key\": \"value\"}\n```"
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Equal(t, `{"key": "value"}`, result)
	})

	t.Run("ブレースで抽出", func(t *testing.T) {
		text := "前置き {\"key\": \"value\"} 後書き"
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Equal(t, `{"key": "value"}`, result)
	})

	t.Run("ネストされた JSON", func(t *testing.T) {
		text := `前置き {"outer": {"inner": "value"}} 後書き`
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Equal(t, `{"outer": {"inner": "value"}}`, result)
	})

	t.Run("json コードブロックが優先される", func(t *testing.T) {
		text := "前置き {\"bad\": true}\n```json\n{\"good\": true}\n```"
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Equal(t, `{"good": true}`, result)
	})

	t.Run("複数行の JSON", func(t *testing.T) {
		text := "```json\n{\n  \"key\": \"value\",\n  \"num\": 42\n}\n```"
		result, err := ExtractJSON(text)
		require.NoError(t, err)
		assert.Contains(t, result, `"key": "value"`)
		assert.Contains(t, result, `"num": 42`)
	})

	t.Run("JSON がない場合はエラー", func(t *testing.T) {
		text := "JSON は含まれていません"
		_, err := ExtractJSON(text)
		assert.Error(t, err)
	})

	t.Run("空文字列はエラー", func(t *testing.T) {
		_, err := ExtractJSON("")
		assert.Error(t, err)
	})

	t.Run("閉じブレースだけはエラー", func(t *testing.T) {
		_, err := ExtractJSON("}")
		assert.Error(t, err)
	})
}
