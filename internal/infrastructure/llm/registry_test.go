package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// テスト用のモッククライアント
type stubClient struct{}

func (s *stubClient) Chat(_ context.Context, _, _ string) (string, error) {
	return "stub", nil
}

func (s *stubClient) ChatWithOptions(_ context.Context, _, _ string, _ ChatOptions) (string, error) {
	return "stub", nil
}

func (s *stubClient) ModelInfo() string {
	return "Stub / stub-model"
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	t.Run("登録したプロバイダのクライアントを取得できる", func(t *testing.T) {
		r := NewRegistry()
		client := &stubClient{}
		r.Register(ProviderOpenAI, client)

		got, err := r.Get(ProviderOpenAI)

		assert.NoError(t, err)
		assert.Equal(t, client, got)
	})

	t.Run("未登録のプロバイダはエラーを返す", func(t *testing.T) {
		r := NewRegistry()

		got, err := r.Get(ProviderClaude)

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "claude")
	})

	t.Run("複数のプロバイダを登録できる", func(t *testing.T) {
		r := NewRegistry()
		openaiClient := &stubClient{}
		claudeClient := &stubClient{}

		r.Register(ProviderOpenAI, openaiClient)
		r.Register(ProviderClaude, claudeClient)

		got1, err := r.Get(ProviderOpenAI)
		assert.NoError(t, err)
		assert.Equal(t, openaiClient, got1)

		got2, err := r.Get(ProviderClaude)
		assert.NoError(t, err)
		assert.Equal(t, claudeClient, got2)
	})

	t.Run("同じプロバイダに再登録すると上書きされる", func(t *testing.T) {
		r := NewRegistry()
		client1 := &stubClient{}
		client2 := &stubClient{}

		r.Register(ProviderOpenAI, client1)
		r.Register(ProviderOpenAI, client2)

		got, err := r.Get(ProviderOpenAI)
		assert.NoError(t, err)
		assert.Equal(t, client2, got)
	})
}

func TestRegistry_Has(t *testing.T) {
	t.Run("登録済みのプロバイダは true を返す", func(t *testing.T) {
		r := NewRegistry()
		r.Register(ProviderOpenAI, &stubClient{})

		assert.True(t, r.Has(ProviderOpenAI))
	})

	t.Run("未登録のプロバイダは false を返す", func(t *testing.T) {
		r := NewRegistry()

		assert.False(t, r.Has(ProviderGemini))
	})
}
