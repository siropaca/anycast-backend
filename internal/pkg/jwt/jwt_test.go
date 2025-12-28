package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenManager_Generate(t *testing.T) {
	manager := NewTokenManager("test-secret")

	t.Run("トークンを生成できる", func(t *testing.T) {
		userID := "user-123"
		expiration := 24 * time.Hour

		token, err := manager.Generate(userID, expiration)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("異なるユーザー ID で異なるトークンが生成される", func(t *testing.T) {
		expiration := 24 * time.Hour

		token1, err := manager.Generate("user-1", expiration)
		require.NoError(t, err)

		token2, err := manager.Generate("user-2", expiration)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})
}

func TestTokenManager_Validate(t *testing.T) {
	manager := NewTokenManager("test-secret")

	t.Run("有効なトークンを検証できる", func(t *testing.T) {
		userID := "user-123"
		token, err := manager.Generate(userID, 24*time.Hour)
		require.NoError(t, err)

		claims, err := manager.Validate(token)

		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, userID, claims.Subject)
	})

	t.Run("期限切れのトークンはエラーになる", func(t *testing.T) {
		userID := "user-123"
		token, err := manager.Generate(userID, -1*time.Hour) // 過去の時刻
		require.NoError(t, err)

		claims, err := manager.Validate(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("不正なトークンはエラーになる", func(t *testing.T) {
		claims, err := manager.Validate("invalid-token")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("異なるシークレットで署名されたトークンはエラーになる", func(t *testing.T) {
		otherManager := NewTokenManager("other-secret")
		token, err := otherManager.Generate("user-123", 24*time.Hour)
		require.NoError(t, err)

		claims, err := manager.Validate(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
